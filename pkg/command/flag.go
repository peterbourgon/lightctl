package command

import "strconv"

type optionalString struct {
	set bool
	s   string
}

func (o *optionalString) Set(s string) error {
	o.set, o.s = true, s
	return nil
}

func (o *optionalString) String() string {
	return o.s
}

type optionalInt struct {
	set bool
	i   int
}

func (o *optionalInt) Set(s string) error {
	i, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	o.set, o.i = true, i
	return nil
}

func (o *optionalInt) String() string {
	return strconv.Itoa(o.i)
}
