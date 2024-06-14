package structs

type Execution_Result struct {
	Gas_Used        int
	Return_Value    interface{}
	Error           error
}