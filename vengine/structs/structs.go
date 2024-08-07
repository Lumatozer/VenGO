package structs

type Execution_Result struct {
	Gas_Used        int
	Return_Value    interface{}
	Error           error
}

type Package struct {
	Name      string
	Functions         map[string]func([]*interface{})Execution_Result
}