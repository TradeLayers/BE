package explicitinit

var packageName string // want `variable "packageName" must be explicitly initialized`
var packageCount int = 0
var packageTotal = 0 // want `variable "packageTotal" must declare an explicit type`

var (
	groupName  string // want `variable "groupName" must be explicitly initialized`
	groupSet   string = ""
	groupTotal        = 0 // want `variable "groupTotal" must declare an explicit type`
)

func valid() {
	var name string = ""
	var ICount int = 0

	_ = name
	_ = ICount
}

func invalid() {
	var missing string // want `variable "missing" must be explicitly initialized`
	var ICount int     // want `variable "ICount" must be explicitly initialized`
	var inferred = 0   // want `variable "inferred" must declare an explicit type`

	_ = missing
	_ = ICount
	_ = inferred
}
