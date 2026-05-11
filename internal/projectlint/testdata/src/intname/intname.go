package intname

func valid(IValue int) {
	var ITotal int = 0
	var Itemp float32
	var IShort int
	for IIndex := range []string{"a"} {
		_ = IIndex
	}

	_ = IValue
	_ = ITotal
	_ = IShort
	_ = Itemp
}

func invalid(count int) { // want `int variable "count" must start with I`
	var total int = 0        // want `int variable "total" must start with I`
	value := 1               // want `int variable "value" must start with I`
	for i := 0; i < 1; i++ { // want `int variable "i" must start with I`
		_ = i
	}
	for idx := range []string{"a"} { // want `int variable "idx" must start with I`
		_ = idx
	}

	_ = count
	_ = total
	_ = value
}
