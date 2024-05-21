package main

func str_index_in_str_arr(a string, b []string) int {
	for i:=0; i<len(b); i++ {
		if b[i]==a {
			return i
		}
	}
	return -1
}