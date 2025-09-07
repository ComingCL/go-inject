package inject

import (
	"bytes"
	"fmt"
	"math/rand"
	"reflect"
)

type Logger interface {
	Info(msg string, keysAndValues ...any)
}

// Populate 是一个便捷函数，用于使用给定的不完整对象值填充依赖图
func Populate(values ...interface{}) error {
	var g Graph
	for _, v := range values {
		if err := g.Provide(&Object{Value: v}); err != nil {
			return err
		}
	}
	return g.Populate()
}

// Object Graph中的一个对象
type Object struct {
	Value        interface{}
	Name         string             // 可选的名称
	Complete     bool               // 如果为true，该Value将被视为完整的
	Fields       map[string]*Object // 填充已注入的字段名称及其对应的*Object
	reflectType  reflect.Type
	reflectValue reflect.Value
	private      bool // 如果为true，该Value将不会被使用，只会被填充
	created      bool // 如果为true，该Object是由我们创建的
	embedded     bool // 如果为true，该Object是内部提供的嵌入结构体
}

func (o *Object) String() string {
	var buf bytes.Buffer
	fmt.Fprint(&buf, o.reflectType)
	if o.Name != "" {
		fmt.Fprintf(&buf, " named %s", o.Name)
	}
	return buf.String()
}

func (o *Object) addDep(field string, dep *Object) {
	if o.Fields == nil {
		o.Fields = make(map[string]*Object)
	}
	o.Fields[field] = dep
}

type Graph struct {
	Logger      Logger // 可选的，将触发信息日志
	unnamed     []*Object
	unnamedType map[reflect.Type]bool
	named       map[string]*Object
}

func (g *Graph) Provide(objects ...*Object) error {
	for _, o := range objects {
		o.reflectType = reflect.TypeOf(o.Value)
		o.reflectValue = reflect.ValueOf(o.Value)

		if o.Fields != nil {
			return fmt.Errorf("fields were specified on object %v when it was provided", o)
		}

		if o.Name == "" {
			if !isStructPtr(o.reflectType) {
				return fmt.Errorf(
					"expected unnamed object value to be a pointer to a struct but got type %s with value %v",
					o.reflectType,
					o.Value)
			}

			if !o.private {
				if g.unnamedType == nil {
					g.unnamedType = make(map[reflect.Type]bool)
				}

				if g.unnamedType[o.reflectType] {
					return fmt.Errorf(
						"provided two unnamed instances of type *%s.%s",
						o.reflectType.Elem().PkgPath(),
						o.reflectType.Elem().Name(),
					)
				}
				g.unnamedType[o.reflectType] = true
			}
			g.unnamed = append(g.unnamed, o)
		} else {
			if g.named == nil {
				g.named = make(map[string]*Object)
			}

			if g.named[o.Name] != nil {
				return fmt.Errorf("provided two instances named %s", o.Name)
			}
			g.named[o.Name] = o
		}

		if g.Logger != nil {
			if o.created {
				g.Logger.Info("created %v", o)
			} else if o.embedded {
				g.Logger.Info("provided embedded %v", o)
			} else {
				g.Logger.Info("provided %v", o)
			}
		}
	}
	return nil
}

// Populate 填充不完整的对象
func (g *Graph) Populate() error {
	for _, o := range g.named {
		if o.Complete {
			continue
		}

		if err := g.populateExplicit(o); err != nil {
			return err
		}
	}

	// 我们在遍历过程中会追加和修改切片，所以不使用标准的
	// range循环，而是对图中的每个对象进行单次遍历。
	i := 0
	for {
		if i == len(g.unnamed) {
			break
		}

		o := g.unnamed[i]
		i++

		if o.Complete {
			continue
		}

		if err := g.populateExplicit(o); err != nil {
			return err
		}
	}

	// 第二遍处理接口值的注入，以确保我们首先创建了所有具体类型。
	for _, o := range g.unnamed {
		if o.Complete {
			continue
		}

		if err := g.populateUnnamedInterface(o); err != nil {
			return err
		}
	}

	for _, o := range g.named {
		if o.Complete {
			continue
		}

		if err := g.populateUnnamedInterface(o); err != nil {
			return err
		}
	}

	return nil
}

func (g *Graph) populateExplicit(o *Object) error {
	// 忽略命名的值类型
	if o.Name != "" && !isStructPtr(o.reflectType) {
		return nil
	}

StructLoop:
	for i := 0; i < o.reflectValue.Elem().NumField(); i++ {
		field := o.reflectValue.Elem().Field(i)
		fieldType := field.Type()
		fieldTag := o.reflectType.Elem().Field(i).Tag
		fieldName := o.reflectType.Elem().Field(i).Name
		tag, err := parseTag(string(fieldTag))
		if err != nil {
			return fmt.Errorf(
				"unexpected tag format `%s` for field %s in type %s",
				string(fieldTag),
				o.reflectType.Elem().Field(i).Name,
				o.reflectType,
			)
		}

		// 跳过没有标签的字段。
		if tag == nil {
			continue
		}

		// 不能用于未导出的字段。
		if !field.CanSet() {
			return fmt.Errorf(
				"inject requested on unexported field %s in type %s",
				o.reflectType.Elem().Field(i).Name,
				o.reflectType,
			)
		}

		// 在结构体以外的任何类型上使用inline标签都被认为是无效的。
		if tag.Inline && fieldType.Kind() != reflect.Struct {
			return fmt.Errorf(
				"inline requested on non inlined field %s in type %s",
				o.reflectType.Elem().Field(i).Name,
				o.reflectType,
			)
		}

		// 不要覆盖现有值，但检查现有值是否需要深度注入。
		if !isNilOrZero(field, fieldType) {
			// 如果字段已经有值且是结构体指针，检查是否需要深度注入
			if isStructPtr(fieldType) && !tag.Private {
				existingValue := field.Interface()
				// 检查这个对象是否已经在依赖图中
				var found bool
				for _, existing := range g.unnamed {
					if existing.Value == existingValue {
						found = true
						break
					}
				}
				// 如果不在依赖图中，添加并递归注入
				if !found {
					existingObject := &Object{
						Value:   existingValue,
						private: false,
						created: false,
					}
					if err := g.Provide(existingObject); err == nil {
						// 递归填充现有对象的依赖（深度注入）
						if err := g.populateExplicit(existingObject); err != nil {
							return err
						}
						if g.Logger != nil {
							g.Logger.Info("deep injected existing %v in field %s of %v", existingObject, o.reflectType.Elem().Field(i).Name, o)
						}
					}
				}
			}
			continue
		}

		// 命名注入必须已经明确提供。
		if tag.Name != "" {
			existing := g.named[tag.Name]
			if existing == nil {
				return fmt.Errorf(
					"did not find object named %s required by field %s in type %s",
					tag.Name,
					o.reflectType.Elem().Field(i).Name,
					o.reflectType,
				)
			}

			if !existing.reflectType.AssignableTo(fieldType) {
				return fmt.Errorf(
					"object named %s of type %s is not assignable to field %s (%s) in type %s",
					tag.Name,
					fieldType,
					o.reflectType.Elem().Field(i).Name,
					existing.reflectType,
					o.reflectType,
				)
			}

			field.Set(reflect.ValueOf(existing.Value))
			if g.Logger != nil {
				g.Logger.Info("assigned %v to field %s in %v", existing, o.reflectType.Elem().Field(i).Name, o)
			}
			o.addDep(fieldName, existing)
			continue StructLoop
		}

		// 内联结构体值表示我们想要遍历进入它，但不注入它本身。
		// 我们需要一个明确的"inline"标签来使其工作
		if fieldType.Kind() == reflect.Struct {
			if tag.Private {
				return fmt.Errorf(
					"cannot use private inject on inline struct on field %s in type %s",
					o.reflectType.Elem().Field(i).Name,
					o.reflectType,
				)
			}

			if !tag.Inline {
				return fmt.Errorf(
					"inline struct on field %s in type %s required an explicit \"inline\" tag",
					o.reflectType.Elem().Field(i).Name,
					o.reflectType,
				)
			}

			if err = g.Provide(&Object{
				Value:    field.Addr().Interface(),
				private:  true,
				embedded: o.reflectType.Elem().Field(i).Anonymous,
			}); err != nil {
				return err
			}
			continue
		}

		// 接口注入在第二遍中处理
		if fieldType.Kind() == reflect.Interface {
			continue
		}

		// Map被创建并且必须是私有的
		if fieldType.Kind() == reflect.Map {
			if !tag.Private {
				return fmt.Errorf(
					"inject on map field %s in type %s must be named or private",
					o.reflectType.Elem().Field(i).Name,
					o.reflectType,
				)
			}

			field.Set(reflect.MakeMap(fieldType))
			if g.Logger != nil {
				g.Logger.Info("made map for field %s in %v", o.reflectType.Elem().Field(i).Name, o)
			}
			continue
		}

		// 从这里开始只能注入指针。
		if !isStructPtr(fieldType) {
			return fmt.Errorf(
				"found inject tag on unsupported field %s in type %s",
				o.reflectType.Elem().Field(i).Name,
				o.reflectType,
			)
		}

		// 除非是私有注入，否则我们将寻找相同类型的现有实例。
		if !tag.Private {
			for _, existing := range g.unnamed {
				if existing.private {
					continue
				}
				if existing.reflectType.AssignableTo(fieldType) {
					field.Set(reflect.ValueOf(existing.Value))
					if g.Logger != nil {
						g.Logger.Info("assigned existing %v to field %s in %v", existing, o.reflectType.Elem().Field(i).Name, o)
					}
					o.addDep(fieldName, existing)
					continue StructLoop
				}
			}
		}

		newValue := reflect.New(fieldType.Elem())
		newObject := &Object{
			Value:   newValue.Interface(),
			private: tag.Private,
			created: true,
		}

		// 将新创建的对象添加到已知对象集合中。
		if err = g.Provide(newObject); err != nil {
			return err
		}

		// 递归填充新创建对象的依赖（深度注入）
		if err = g.populateExplicit(newObject); err != nil {
			return err
		}

		// 最后将新创建的对象分配给我们的字段。
		field.Set(newValue)
		if g.Logger != nil {
			g.Logger.Info("assigned newly created %v to field %s in %v", newObject, o.reflectType.Elem().Field(i).Name, o)
		}
		o.addDep(fieldName, newObject)
	}
	return nil
}

func (g *Graph) populateUnnamedInterface(o *Object) error {
	// 忽略命名的值类型.
	if o.Name != "" && !isStructPtr(o.reflectType) {
		return nil
	}

	for i := 0; i < o.reflectValue.Elem().NumField(); i++ {
		field := o.reflectValue.Elem().Field(i)
		fieldType := field.Type()
		fieldTag := o.reflectType.Elem().Field(i).Tag
		fieldName := o.reflectType.Elem().Field(i).Name
		tag, err := parseTag(string(fieldTag))
		if err != nil {
			return fmt.Errorf(
				"unexpected tag format `%s` for field %s in type %s",
				string(fieldTag),
				o.reflectType.Elem().Field(i).Name,
				o.reflectType,
			)
		}

		// 跳过没有标签的字段。
		if tag == nil {
			continue
		}

		// 我们在这里只处理接口注入。其他情况包括错误
		// 在第一遍注入指针时处理
		if fieldType.Kind() != reflect.Interface {
			continue
		}

		// 接口注入不能是私有的，因为我们无法实例化接口的新实例。
		if tag.Private {
			return fmt.Errorf(
				"found private inject tag on interface field %s in type %s",
				o.reflectType.Elem().Field(i).Name,
				o.reflectType,
			)
		}

		// 不要覆盖现有值。
		if !isNilOrZero(field, fieldType) {
			continue
		}

		// 命名注入必须已经在populateExplicit中处理。
		if tag.Name != "" {
			panic(fmt.Sprintf("unhandled named instance with name %s", tag.Name))
		}

		// 为字段找到一个且仅一个可分配的值。
		var found *Object
		for _, existing := range g.unnamed {
			if existing.private {
				continue
			}
			if existing.reflectType.AssignableTo(fieldType) {
				if found != nil {
					return fmt.Errorf(
						"found two assignable values for field %s in type %s. one type %s with value %v and another type %s with value %v",
						o.reflectType.Elem().Field(i).Name,
						o.reflectType,
						found.reflectType,
						found.Value,
						existing.reflectType,
						existing.reflectValue,
					)
				}
				found = existing
				field.Set(reflect.ValueOf(existing.Value))
				if g.Logger != nil {
					g.Logger.Info("assigned existing %v to interface field %s in %v", existing, o.reflectType.Elem().Field(i).Name, o)
				}
				o.addDep(fieldName, existing)
			}
		}
		if found == nil {
			return fmt.Errorf("found no assignable value for field %s in type %s",
				o.reflectType.Elem().Field(i).Name,
				o.reflectType,
			)
		}
	}

	return nil
}

// Objects 返回所有已知对象，包括命名的和未命名的。返回的
// 元素不是稳定顺序的。
func (g *Graph) Objects() []*Object {
	objects := make([]*Object, 0, len(g.unnamed)+len(g.named))
	for _, o := range g.unnamed {
		if !o.embedded {
			objects = append(objects, o)
		}
	}
	for _, o := range g.named {
		if !o.embedded {
			objects = append(objects, o)
		}
	}
	// 随机化以防止调用者依赖排序
	for i := 0; i < len(objects); i++ {
		j := rand.Intn(i + 1)
		objects[i], objects[j] = objects[j], objects[i]
	}
	return objects
}

var (
	injectOnly    = &tag{}
	injectPrivate = &tag{Private: true}
	injectInline  = &tag{Inline: true}
)

type tag struct {
	Name    string
	Inline  bool
	Private bool
}

func parseTag(t string) (*tag, error) {
	found, value, err := Extract("inject", t)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	if value == "" {
		return injectOnly, nil
	}
	if value == "inline" {
		return injectInline, nil
	}
	if value == "private" {
		return injectPrivate, nil
	}
	return &tag{Name: value}, nil
}

func isStructPtr(t reflect.Type) bool {
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
}

func isNilOrZero(v reflect.Value, t reflect.Type) bool {
	switch v.Kind() {
	default:
		return reflect.DeepEqual(v.Interface(), reflect.Zero(t).Interface())
	}
}
