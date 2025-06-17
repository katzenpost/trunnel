package gen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/katzenpost/trunnel/ast"
	"github.com/katzenpost/trunnel/inspect"
)

// Marshallers builds data marshallers for types in the given files.
func Marshallers(pkg string, fs []*ast.File) ([]byte, error) {
	g := &generator{}
	if err := g.files(pkg, fs); err != nil {
		return nil, err
	}
	return g.imported()
}

type generator struct {
	printer

	resolver *inspect.Resolver

	receiver string // method receiver variable
	data     string // data variable
}

func (g *generator) files(pkg string, fs []*ast.File) error {
	g.header(pkg)

	if err := g.init(fs); err != nil {
		return err
	}

	for _, f := range fs {
		if err := g.file(f); err != nil {
			return err
		}
	}

	return nil
}

func (g *generator) init(fs []*ast.File) (err error) {
	g.resolver, err = inspect.NewResolverFiles(fs)
	return
}

func (g *generator) file(f *ast.File) error {
	for _, c := range f.Contexts {
		g.context(c)
	}

	for _, s := range f.Structs {
		g.structure(s)
	}

	return nil
}

func (g *generator) context(c *ast.Context) {
	g.printf("type %s struct {\n", name(c.Name))
	for _, m := range c.Members {
		g.structMemberDecl(m)
	}
	g.printf("}\n\n")
}

func (g *generator) structure(s *ast.Struct) {
	if s.Extern() {
		return
	}

	g.receiver = strings.ToLower(s.Name[:1])

	g.structDecl(s)
	g.parse(s)
	g.parseConstructor(s)
	g.marshalBinary(s)

	g.receiver = ""
}

func (g *generator) structDecl(s *ast.Struct) {
	g.printf("type %s struct {\n", name(s.Name))
	for _, m := range s.Members {
		g.structMemberDecl(m)
	}
	g.printf("}\n\n")
}

func (g *generator) structMemberDecl(m ast.Member) {
	switch m := m.(type) {
	case *ast.Field:
		g.printf("\t%s %s\n", name(m.Name), g.tipe(m.Type))
	case *ast.UnionMember:
		g.structUnionMemberDecl(m)
	case *ast.EOS:
		// ignore
	default:
		panic(unexpected(m))
	}
}

func (g *generator) structUnionMemberDecl(m *ast.UnionMember) {
	for _, c := range m.Cases {
		for _, f := range c.Members {
			switch f := f.(type) {
			case *ast.Fail, *ast.Ignore:
				// nothing
			default:
				g.structMemberDecl(f)
			}
		}
	}
}

func (g *generator) parseConstructor(s *ast.Struct) {
	n := name(s.Name)
	g.printf("func Parse%s(data []byte%s) (*%s, error) {\n", n, contextSignature(s.Contexts), n)
	g.printf("%s := new(%s)\n", g.receiver, n)
	g.printf("_, err := %s.Parse(data%s)\n", g.receiver, contextArgs(s.Contexts))
	g.printf("if err != nil { return nil, err }\n")
	g.printf("return %s, nil\n", g.receiver)
	g.printf("}\n\n")
}

// parse generates a parse function for the type.
func (g *generator) parse(s *ast.Struct) {
	g.printf("func (%s *%s) Parse(data []byte%s) ([]byte, error) {\n",
		g.receiver, name(s.Name), contextSignature(s.Contexts))
	g.printf("cur := data\n")
	g.data = "cur"
	for _, m := range s.Members {
		g.parseMember(m)
	}
	g.printf("return %s, nil\n}\n\n", g.data)
	g.data = ""
}

func (g *generator) parseMember(m ast.Member) {
	g.printf("{\n")
	switch m := m.(type) {
	case *ast.Field:
		lhs := g.receiver + "." + name(m.Name)
		g.parseType(lhs, m.Type)

	case *ast.UnionMember:
		g.parseUnionMember(m)

	case *ast.EOS:
		g.assertEnd()

	case *ast.Ignore:
		g.printf("%s = []byte{}\n", g.data)

	case *ast.Fail:
		g.printf("return nil, errors.New(\"disallowed case\")")

	default:
		panic(unexpected(m))
	}
	g.printf("}\n")
}

func (g *generator) parseType(lhs string, t ast.Type) {
	switch t := t.(type) {
	case *ast.NulTermString:
		g.printf("i := bytes.IndexByte(%s, 0)\n", g.data)
		g.printf("if i < 0 { return nil, errors.New(\"could not parse nul-term string\") }\n")
		g.printf("%s, %s = string(%s[:i]), %s[i+1:]\n", lhs, g.data, g.data, g.data)

	case *ast.IntType:
		g.parseIntType(lhs, t)

	case *ast.CharType:
		g.parseType(lhs, ast.U8)

	case *ast.Ptr:
		g.printf("%s = len(data) - len(%s)\n", lhs, g.data)

	case *ast.StructRef:
		g.printf("var err error\n")
		g.printf("%s = new(%s)\n", lhs, name(t.Name))
		s, ok := g.resolver.Struct(t.Name)
		if !ok {
			panic("struct not found") // XXX return err
		}
		g.printf("%s, err = %s.Parse(%s%s)\n", g.data, lhs, g.data, contextArgs(s.Contexts))
		g.printf("if err != nil { return nil, err }\n")

	case *ast.FixedArrayMember:
		g.parseArray(lhs, t.Base, t.Size)

	case *ast.VarArrayMember:
		g.parseArray(lhs, t.Base, t.Constraint)

	default:
		panic(unexpected(t))
	}
}

func (g *generator) parseIntType(lhs string, t *ast.IntType) {
	n := t.Size / 8
	g.lengthCheck(strconv.Itoa(int(n)))
	if n == 1 {
		g.printf("%s = %s[0]\n", lhs, g.data)
	} else {
		g.printf("%s = binary.BigEndian.Uint%d(%s)\n", lhs, t.Size, g.data)
	}
	if t.Constraint != nil {
		g.printf("if !(%s) {\n", g.conditional(lhs, t.Constraint))
		g.printf("return nil, errors.New(\"integer constraint violated\")\n")
		g.printf("}\n")
	}
	g.printf("%s = %s[%d:]\n", g.data, g.data, n)
}

func (g *generator) parseArray(lhs string, base ast.Type, s ast.LengthConstraint) {
	switch s := s.(type) {
	case *ast.IntegerConstRef, *ast.IntegerLiteral:
		g.printf("for idx := 0; idx < %s; idx++ {\n", g.integer(s))
		g.parseType(lhs+"[idx]", base)
		g.printf("}\n")

	case *ast.IDRef:
		size := fmt.Sprintf("int(%s)", g.ref(s))
		g.printf("%s = make([]%s, %s)\n", lhs, g.tipe(base), size)
		g.printf("for idx := 0; idx < %s; idx++ {\n", size)
		g.parseType(lhs+"[idx]", base)
		g.printf("}\n")

	case *ast.Leftover:
		g.constrained(s, func() {
			g.parseArray(lhs, base, nil)
		})

	case nil:
		g.printf("%s = make([]%s, 0)\n", lhs, g.tipe(base))
		g.printf("for len(%s) > 0 {\n", g.data)
		g.printf("var tmp %s\n", g.tipe(base))
		g.parseType("tmp", base)
		g.printf("%s = append(%s, tmp)\n", lhs, lhs)
		g.printf("}\n")

	default:
		panic(unexpected(s))
	}
}

func (g *generator) parseUnionMember(u *ast.UnionMember) {
	if u.Length != nil {
		g.constrained(u.Length, func() {
			g.parseUnionMember(&ast.UnionMember{
				Name:  u.Name,
				Tag:   u.Tag,
				Cases: u.Cases,
			})
		})
		return
	}

	tag := g.ref(u.Tag)
	g.printf("switch {\n")
	for _, c := range u.Cases {
		if c.Case == nil {
			g.printf("default:\n")
		} else {
			g.printf("case %s:\n", g.conditional(tag, c.Case))
		}
		for _, m := range c.Members {
			g.parseMember(m)
		}
	}
	g.printf("}\n")
}

func (g *generator) constrained(c ast.LengthConstraint, f func()) {
	var n string

	switch c := c.(type) {
	case *ast.Leftover:
		g.lengthCheck(g.integer(c.Num))
		n = fmt.Sprintf("len(%s)-%s", g.data, g.integer(c.Num))

	case *ast.IDRef:
		n = fmt.Sprintf("int(%s)", g.ref(c))
		g.lengthCheck(n)

	default:
		panic(unexpected(c))
	}

	g.printf("restore := %s[%s:]\n", g.data, n)
	g.printf("%s = %s[:%s]\n", g.data, g.data, n)
	f()
	g.assertEnd()
	g.printf("%s = restore\n", g.data)
}

// ref builds a variable reference that resolves to the given trunnel IDRef.
func (g *generator) ref(r *ast.IDRef) string {
	if r.Scope == "" {
		return g.receiver + "." + name(r.Name)
	}
	return r.Scope + "." + name(r.Name)
}

func (g *generator) lengthCheck(min string) {
	g.printf("if len(%s) < %s { return nil, errors.New(\"data too short\") }\n", g.data, min)
}

func (g *generator) assertEnd() {
	g.printf("if len(%s) > 0 { return nil, errors.New(\"trailing data disallowed\") }\n", g.data)
}

func (g *generator) integer(i ast.Integer) string {
	x, err := g.resolver.Integer(i)
	if err != nil {
		panic(err) // XXX panic
	}
	return strconv.FormatInt(x, 10)
}

func (g *generator) conditional(v string, c *ast.IntegerList) string {
	clauses := make([]string, len(c.Ranges))
	for i, r := range c.Ranges {
		// Single case
		if r.High == nil {
			clauses[i] = fmt.Sprintf("%s == %s", v, g.integer(r.Low))
		} else {
			clauses[i] = fmt.Sprintf("(%s <= %s && %s <= %s)", g.integer(r.Low), v, v, g.integer(r.High))
		}
	}
	return strings.Join(clauses, " || ")
}

func (g *generator) tipe(t interface{}) string {
	switch t := t.(type) {
	case *ast.NulTermString:
		return "string"
	case *ast.IntType:
		return fmt.Sprintf("uint%d", t.Size)
	case *ast.CharType:
		return "byte"
	case *ast.Ptr:
		return "int"
	case *ast.StructRef:
		return "*" + name(t.Name)
	case *ast.FixedArrayMember:
		return fmt.Sprintf("[%s]%s", g.integer(t.Size), g.tipe(t.Base))
	case *ast.VarArrayMember:
		return fmt.Sprintf("[]%s", g.tipe(t.Base))
	default:
		panic(unexpected(t))
	}
}

func contextSignature(names []string) string {
	s := ""
	for _, n := range names {
		s += ", " + n + " " + name(n)
	}
	return s
}

func contextArgs(names []string) string {
	s := ""
	for _, n := range names {
		s += ", " + n
	}
	return s
}

func unexpected(t interface{}) string {
	return fmt.Sprintf("unexpected type %T", t)
}

// marshalBinary generates a MarshalBinary method for the type.
func (g *generator) marshalBinary(s *ast.Struct) {
	// Generate internal encoding method without validation
	g.printf("func (%s *%s) encodeBinary() []byte {\n", g.receiver, name(s.Name))
	g.printf("var buf []byte\n")
	g.data = "buf"
	for _, m := range s.Members {
		g.encodeMember(m)
	}
	g.printf("return %s\n}\n\n", g.data)
	g.data = ""

	// Generate public MarshalBinary method with validation
	g.printf("func (%s *%s) MarshalBinary() ([]byte, error) {\n", g.receiver, name(s.Name))
	g.printf("if err := %s.validate(); err != nil {\n", g.receiver)
	g.printf("return nil, err\n")
	g.printf("}\n")
	g.printf("return %s.encodeBinary(), nil\n", g.receiver)
	g.printf("}\n\n")

	// Generate validation method
	g.validate(s)
}

func (g *generator) encodeMember(m ast.Member) {
	switch m := m.(type) {
	case *ast.Field:
		rhs := g.receiver + "." + name(m.Name)
		g.encodeType(rhs, m.Type)

	case *ast.UnionMember:
		g.encodeUnionMember(m)

	case *ast.EOS:
		// nothing to encode

	case *ast.Ignore:
		// nothing to encode

	case *ast.Fail:
		// nothing to encode

	default:
		panic(unexpected(m))
	}
}

func (g *generator) encodeType(rhs string, t ast.Type) {
	switch t := t.(type) {
	case *ast.NulTermString:
		g.printf("%s = append(%s, []byte(%s)...)\n", g.data, g.data, rhs)
		g.printf("%s = append(%s, 0)\n", g.data, g.data)

	case *ast.IntType:
		g.encodeIntType(rhs, t)

	case *ast.CharType:
		g.encodeType(rhs, ast.U8)

	case *ast.Ptr:
		// Pointers are computed during parsing, not encoded
		// nothing to encode

	case *ast.StructRef:
		g.printf("%s = append(%s, %s.encodeBinary()...)\n", g.data, g.data, rhs)

	case *ast.FixedArrayMember:
		g.encodeArray(rhs, t.Base, t.Size)

	case *ast.VarArrayMember:
		g.encodeArray(rhs, t.Base, t.Constraint)

	default:
		panic(unexpected(t))
	}
}

func (g *generator) encodeIntType(rhs string, t *ast.IntType) {
	n := t.Size / 8
	if n == 1 {
		g.printf("%s = append(%s, byte(%s))\n", g.data, g.data, rhs)
	} else {
		g.printf("{\n")
		g.printf("tmp := make([]byte, %d)\n", n)
		g.printf("binary.BigEndian.PutUint%d(tmp, %s)\n", t.Size, rhs)
		g.printf("%s = append(%s, tmp...)\n", g.data, g.data)
		g.printf("}\n")
	}
}

func (g *generator) encodeArray(rhs string, base ast.Type, s ast.LengthConstraint) {
	switch s := s.(type) {
	case *ast.IntegerConstRef, *ast.IntegerLiteral:
		g.printf("for idx := 0; idx < %s; idx++ {\n", g.integer(s))
		g.encodeType(rhs+"[idx]", base)
		g.printf("}\n")

	case *ast.IDRef:
		size := fmt.Sprintf("int(%s)", g.ref(s))
		g.printf("for idx := 0; idx < %s; idx++ {\n", size)
		g.encodeType(rhs+"[idx]", base)
		g.printf("}\n")

	case *ast.Leftover:
		// For leftover arrays, encode all elements
		g.printf("for idx := 0; idx < len(%s); idx++ {\n", rhs)
		g.encodeType(rhs+"[idx]", base)
		g.printf("}\n")

	case nil:
		// Variable length array - encode all elements
		g.printf("for idx := 0; idx < len(%s); idx++ {\n", rhs)
		g.encodeType(rhs+"[idx]", base)
		g.printf("}\n")

	default:
		panic(unexpected(s))
	}
}

func (g *generator) encodeUnionMember(u *ast.UnionMember) {
	tag := g.ref(u.Tag)
	g.printf("switch {\n")
	for _, c := range u.Cases {
		if c.Case == nil {
			g.printf("default:\n")
		} else {
			g.printf("case %s:\n", g.conditional(tag, c.Case))
		}
		for _, m := range c.Members {
			g.encodeMember(m)
		}
	}
	g.printf("}\n")
}

func (g *generator) validate(s *ast.Struct) {
	g.printf("func (%s *%s) validate() error {\n", g.receiver, name(s.Name))
	for _, m := range s.Members {
		g.validateMember(m)
	}
	g.printf("return nil\n}\n\n")
}

func (g *generator) validateMember(m ast.Member) {
	switch m := m.(type) {
	case *ast.Field:
		rhs := g.receiver + "." + name(m.Name)
		g.validateType(rhs, m.Type)

	case *ast.UnionMember:
		g.validateUnionMember(m)

	case *ast.EOS, *ast.Ignore, *ast.Fail:
		// nothing to validate

	default:
		panic(unexpected(m))
	}
}

func (g *generator) validateType(rhs string, t ast.Type) {
	switch t := t.(type) {
	case *ast.IntType:
		g.validateIntType(rhs, t)

	case *ast.StructRef:
		g.printf("if %s != nil {\n", rhs)
		g.printf("if err := %s.validate(); err != nil {\n", rhs)
		g.printf("return err\n")
		g.printf("}\n")
		g.printf("}\n")

	case *ast.FixedArrayMember:
		g.validateArray(rhs, t.Base, t.Size)

	case *ast.VarArrayMember:
		g.validateArray(rhs, t.Base, t.Constraint)

	case *ast.NulTermString, *ast.CharType, *ast.Ptr:
		// No validation needed for these types

	default:
		panic(unexpected(t))
	}
}

func (g *generator) validateIntType(rhs string, t *ast.IntType) {
	if t.Constraint != nil {
		g.printf("if !(%s) {\n", g.conditional(rhs, t.Constraint))
		g.printf("return errors.New(\"integer constraint violated\")\n")
		g.printf("}\n")
	}
}

func (g *generator) validateArray(rhs string, base ast.Type, s ast.LengthConstraint) {
	switch s := s.(type) {
	case *ast.IntegerConstRef, *ast.IntegerLiteral:
		expectedSize := g.integer(s)
		g.printf("if len(%s) != %s {\n", rhs, expectedSize)
		g.printf("return errors.New(\"array length constraint violated\")\n")
		g.printf("}\n")

	case *ast.IDRef:
		expectedSize := fmt.Sprintf("int(%s)", g.ref(s))
		g.printf("if len(%s) != %s {\n", rhs, expectedSize)
		g.printf("return errors.New(\"array length constraint violated\")\n")
		g.printf("}\n")

	case *ast.Leftover, nil:
		// No length validation for leftover or variable arrays

	default:
		panic(unexpected(s))
	}

	// Validate array elements
	g.printf("for idx := 0; idx < len(%s); idx++ {\n", rhs)
	g.validateType(rhs+"[idx]", base)
	g.printf("}\n")
}

func (g *generator) validateUnionMember(u *ast.UnionMember) {
	tag := g.ref(u.Tag)
	g.printf("switch {\n")
	for _, c := range u.Cases {
		if c.Case == nil {
			g.printf("default:\n")
		} else {
			g.printf("case %s:\n", g.conditional(tag, c.Case))
		}
		for _, m := range c.Members {
			g.validateMember(m)
		}
	}
	g.printf("}\n")
}
