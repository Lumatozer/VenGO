package main

func str_index_in_str_arr(a string, b []string) int {
	for i:=0; i<len(b); i++ {
		if b[i]==a {
			return i
		}
	}
	return -1
}

func Can_access(index int, arr_len int) bool {
	if index<0 {
		return false
	}
	return arr_len>index
}