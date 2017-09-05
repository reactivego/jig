package pkg

import (
	"log"
	"path/filepath"
)

// ParseDir will add .go files found in the package directory to the internal list of files.
// This will also detect if a file contains previously generated source
func (p *Package) ParseDir() error {

	// Load all go files in folder indicated by path
	filepaths, err := filepath.Glob(filepath.Join(p.Dir, "*.go"))
	if err != nil {
		return err
	}

	// Parse the list of filenames into a list of ast.File objects.
	for _, path := range filepaths {
		file, err := p.Config.ParseFile(path, nil)
		if err != nil {
			log.Fatal(err)
		}

		// Get the package name from the first file that is parsed.
		if p.Name == "" {
			p.Name = file.Name.String()
		}
		// TODO else check that the packageName is the same as the previous one!!

		// Make sure the file is added to the list of files.
		p.AddFile(file)

		// Also scan the file for generated source fragments.
		p.ScanForGeneratedSources(file)
	}

	return nil
}
