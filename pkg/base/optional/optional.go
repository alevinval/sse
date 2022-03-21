package optional

type Optional[T any] struct {
	value   T
	present bool
}

func Empty[T any]() Optional[T] {
	return Optional[T]{present: false}
}

func Of[T any](v T) Optional[T] {
	return Optional[T]{value: v, present: true}
}

func OfNullable[T any](v *T) Optional[T] {
	return Optional[T]{value: *v, present: v != nil}
}

func OfPresent[T any](v T, present bool) Optional[T] {
	return Optional[T]{value: v, present: present}
}

func (o Optional[T]) Get() T {
	if !o.present {
		panic("Get called on an empty optional")
	}
	return o.value
}

func (o Optional[T]) GetOrDefault(d T) T {
	if !o.present {
		return d
	}
	return o.value
}

func (o Optional[T]) IsEmpty() bool {
	return !o.IsPresent()
}

func (o Optional[T]) IsPresent() bool {
	return o.present
}

func (o Optional[T]) IfPresent(fn func(v T)) {
	if o.present {
		fn(o.value)
	}
}
