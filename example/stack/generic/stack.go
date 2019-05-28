package stack

//jig:template <Foo>Stack

type FooStack []foo

var zeroFoo foo

//jig:template <Foo>Stack Push

func (s *FooStack) Push(v foo) {
	*s = append(*s, v)
}

//jig:template <Foo>Stack Pop

func (s *FooStack) Pop() (foo, bool) {
	if len(*s) == 0 {
		return zeroFoo, false
	}
	i := len(*s) - 1
	v := (*s)[i]
	*s = (*s)[:i]
	return v, true
}

//jig:template <Foo>Stack Top

func (s *FooStack) Top() (foo, bool) {
	if len(*s) == 0 {
		return zeroFoo, false
	}
	return (*s)[len(*s)-1], true
}
