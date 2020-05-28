package pkg

// Check will typecheck the currently parsed package source and return all errors
// that were found. This will also import and parse all dependencies.
// After Check() has finished the package contains the contents of all imported
// packages and therefore LoadGenerics() may be caled to locate and load templates
// from those imported packages.
func (p *Package) Check() ([]error, error) {
	//d := time.Now()

	p.CreatePkgs = p.PkgSpec()

	// Load the program, type checking it in the process.
	prog, err := p.Load()
	if err != nil {
		return nil, err
	}

	// Collect all errors that were found into a single slice.
	var errs []error
	for _, created := range prog.Created {
		errs = append(errs, created.Errors...)
	}

	// Append all PackageInfo structs into allPackages.
	p.allPackages = nil
	for _, pkg := range prog.AllPackages {
		p.allPackages = append(p.allPackages, pkg)
	}

	//fmt.Println("Check", time.Since(d))
	return errs, nil
}
