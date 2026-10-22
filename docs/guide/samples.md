# Code Samples and cURL

A walkthrough shows each request in three ways: as HTTP, as a `curl` command, and as the code a developer writes to send or handle it. PIDL generates the HTTP and cURL forms from a step's example. Code samples come from the implementations of the protocol, preferably from their tested source files, so the code a reader sees compiles and runs.

In the [walkthrough viewer](viewer.md), each request and response panel has a tab for HTTP, cURL and every sample language. The reader's choice is kept from step to step; a step without that form shows HTTP with the tab disabled.

## cURL

Every example request gets an equivalent `curl` command: the method, the URL, each header in written order, and the body as it goes on the wire. You write nothing for it.

```bash
curl 'https://api.example/reports' \
  -H 'Authorization: Bearer eyJ0eXAiOiJhdCtqd3Qi…'
```

- **Body.** A body is written as JSON for readability. On the wire it is compact JSON with keys in the order written, which is also what a `Content-Digest` in the example covers. A body written as a JSON string is sent as the string's contents, for media types that are not JSON, such as a form.
- **Signed requests.** curl cannot create HTTP message signatures, so the command for a signed request replays the example's signature headers, and a comment at the top says so.

## Code Samples

A sample is code in one language for one step:

| Field | Meaning |
|-------|---------|
| `language` | The highlighter name, such as `go`, `typescript`, `python` or `csharp`. Not `http` or `curl`, which PIDL generates |
| `label` | Tells samples in one language apart, such as the client's code and the API's |
| `role` | The entity whose developer writes this code |
| `code` | Inline code |
| `source` | Where the code lives: `repo`, optional `ref` (a tag or commit), `path` inside the repository with forward slashes, and `region` |

A step where nothing goes over HTTP, such as a person approving in a browser, can still show code: the servers and the client run code for it.

Samples can live in two places:

- **A sample set**, a file kept by an implementation, beside its tested code. The protocol document then names no implementation, and each SDK keeps its own samples current. Use this for a protocol with several implementations, or a protocol document shared across projects.
- **Inline in the document**, on a step's example. Use this for a protocol and its one implementation in one repository.

### Sample Sets

A sample set attaches samples to steps by their address. Give the steps you want to show code for an `id`:

```json
{
  "format": "pidl-samples/v1",
  "protocol": "oauth2-api-access",
  "name": "example-sdk-go",
  "samples": [
    {"step": "authorization-code/token-request", "language": "go", "label": "Client", "role": "client",
     "source": {"repo": "example.com/sdk", "path": "walkthrough/token_test.go", "region": "authorization-code.token-request"}},
    {"step": "call-api/api-request", "language": "go", "label": "Client", "role": "client",
     "source": {"repo": "example.com/sdk", "path": "walkthrough/api_test.go", "region": "call-api.request"}}
  ]
}
```

- **Addresses** name a step as written: `<scenario>/<step-id>`, including a step of a fragment, which then shows wherever the fragment is included; `<scenario>` for samples about the whole scenario; or `<step-id>` in a document without scenarios.
- **`protocol`** names the document's protocol ID, so a set is not joined to another protocol by mistake.
- **`name`** names the implementation; a step without an example gets one citing it.

Join one or more sets to the document when generating, one per implementation:

```bash
pidl generate -f viewer --samples sdk-go.pidl-samples.json --samples sdk-ts.pidl-samples.json -o api.html api.json
```

An address that names no step, or a sample that is not resolved, is an error, so a set cannot fall silently behind the protocol. The format's schema is [`schema/samples.schema.json`](https://github.com/grokify/pidl/blob/main/schema/samples.schema.json).

### Inline Samples

Put `samples` on a flow's `example`:

```json
{
  "id": "api-request",
  "from": "client", "to": "api", "action": "api_request",
  "example": {
    "source": "RFC 6750 §2.1",
    "kind": "draft",
    "request": {"method": "GET", "url": "https://api.example/reports"},
    "samples": [
      {"language": "go", "label": "Client", "role": "client",
       "source": {"repo": "example.com/sdk", "ref": "v1.2.0",
                  "path": "walkthrough/api_test.go", "region": "call-api.request"}},
      {"language": "python", "label": "Client", "role": "client",
       "code": "res = client.get(\"https://api.example/reports\")"}
    ]
  }
}
```

### Code from Tested Files

Inline code drifts from the library it calls. A `source` points at a region of a real file instead, ideally a test that runs in that project's CI, so every sample compiles and runs with each release of the code it shows.

Mark the region with `pidl:begin` and `pidl:end` comments, in any comment syntax:

```go
func TestCallAPI(t *testing.T) {
	// ...setup...

	// pidl:begin call-api.request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL+"/reports", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	// pidl:end call-api.request
}
```

The sample is the code between the markers. Marker lines of other regions inside it are dropped, the common indentation is removed, and blank lines at either end are trimmed. Naming regions after the step they show keeps the two easy to match.

### Resolving and Checking

Generation never reads the source repositories or the network. It reads code recorded beforehand:

```bash
# A sample set records the code and its SHA-256 in the set itself
pidl samples resolve sdk-go.pidl-samples.json

# A document records them in a lock file beside it, api.samples.json
pidl samples resolve api.json

# Fail when a source is not resolved, its region is gone, or its code changed
pidl samples check sdk-go.pidl-samples.json
pidl samples check api.json
```

`resolve` reads each repository from a checkout at `<root>/<repo>`, where the root is `--root`, or `$GOPATH/src` (`~/go/src`) by default. `--repo-dir <dir>` reads every source from one checkout instead, which suits a sample set in the repository it describes. With a `ref`, a file is read as it is at that tag or commit.

In an implementation's CI, check its set against the protocol document, so a renamed or removed step is caught:

```bash
pidl samples check --repo-dir . --protocol api.json sdk-go.pidl-samples.json
```

When the code a sample shows changes on purpose, run `pidl samples resolve` again and commit the result.

### Narration for a Representation

A step's narration describes the protocol, but a reader looking at the Go tab may want to hear about the code. A step's `representations` override its narration for `http`, `curl` or a sample language. See [Narration](narration.md#representations).

## Related

- [Walkthrough Viewer](viewer.md)
- [Narration](narration.md)
- [Specification: Sample Sets](../SPECIFICATION.md#sample-sets)
- [CLI Reference: samples](../reference/cli.md#samples)
