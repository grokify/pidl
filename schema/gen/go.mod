module github.com/grokify/pidl/schema/gen

go 1.26.1

require (
	github.com/grokify/pidl v0.0.0
	github.com/invopop/jsonschema v0.14.0
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.3
)

require (
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/buger/jsonparser v1.1.2 // indirect
	github.com/grokify/oscompat v0.8.0 // indirect
	github.com/pb33f/ordered-map/v2 v2.3.1 // indirect
	go.yaml.in/yaml/v4 v4.0.0-rc.2 // indirect
	golang.org/x/text v0.43.0 // indirect
)

// The generator reflects this repository's types, so it always builds
// against the checkout it sits in.
replace github.com/grokify/pidl => ../..
