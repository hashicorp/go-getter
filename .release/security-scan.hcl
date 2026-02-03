container {
	dependencies = true
	osv          = true
	secrets      = true

}

binary {
	secrets      = true
	go_modules   = true
	osv          = true
	oss_index    = false
	nvd          = false
}