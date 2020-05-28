package templ

// byNumVarsAndLength attaches sort methods to a []*Generic instance.
// It sorts first on the number of variables from most to least number of
// variables. Withing each tier sorts generics on length of their name, from
// short to longest.
type byNumVarsAndLength []*Generic

func (a byNumVarsAndLength) Len() int {
	return len(a)
}

func (a byNumVarsAndLength) Less(i, j int) bool {
	if len(a[i].Vars) > len(a[j].Vars) {
		// Generics with more variables are considered less.
		return true
	}
	if len(a[i].Vars) < len(a[j].Vars) {
		// Generics with less variables are considered more
		return false
	}
	// Equal number of variables, then compare length of their Name.
	return len(a[i].Name) < len(a[j].Name)
}

func (a byNumVarsAndLength) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
