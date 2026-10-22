# Code Samples and cURL

A walkthrough shows each request in three ways: as HTTP, as a `curl` command, and as the code a developer writes to send or handle it. PIDL generates the HTTP and cURL forms from a step's example. Code samples you add yourself, preferably from tested source files, so the code a reader sees compiles and runs.

In the [walkthrough viewer](viewer.md), each request and response panel has a tab for HTTP, cURL and every sample language. The reader's choice is kept from step to step; a step without that form shows HTTP with the tab disabled.

## cURL

Every example request gets an equivalent `curl` command: the method, the URL, each header in written order, and the body as it goes on the wire. You write nothing for it.

```bash
curl 'https://resource.example/api/documents' \
  -H 'Signature-Input: sig=("@method" "@authority" "@path" "signature-key");created=1730217610' \
  -H 'Signature: sig=:SIGNATURE-PLACEHOLDER:' \
  -H 'Signature-Key: sig=jwt;jwt="eyJhbGciOi…"'
```

- **Body.** A body is written as JSON for readability. On the wire it is compact JSON with keys in the order written, which is also what a `Content-Digest` in the example covers. A body written as a JSON string is sent as the string's contents, for media types that are not JSON.
- **Signed requests.** curl cannot create HTTP message signatures, so the command for a signed request replays the example's signature headers, and a comment at the top says so.

## Code Samples

A sample is code in one language for one step. Put `samples` on a flow's `example`:

```json
{
  "id": "request",
  "from": "agent", "to": "resource", "action": "signed_request",
  "example": {
    "source": "draft-11 §5",
    "kind": "draft",
    "request": {"method": "GET", "url": "https://resource.example/api/documents"},
    "samples": [
      {"language": "go", "label": "Agent", "role": "agent",
       "source": {"repo": "github.com/aauth-dev/aauth-go", "ref": "v0.3.0",
                  "path": "examples/walkthrough/agent_identity_test.go",
                  "region": "agent-identity.request.agent"}},
      {"language": "python", "label": "Agent", "role": "agent",
       "code": "res = client.get(\"https://resource.example/api/documents\")"}
    ]
  }
}
```

| Field | Meaning |
|-------|---------|
| `language` | The highlighter name, such as `go`, `typescript`, `python` or `csharp`. Not `http` or `curl`, which PIDL generates |
| `label` | Tells samples in one language apart, such as the agent's code and the resource's |
| `role` | The entity whose developer writes this code |
| `code` | Inline code |
| `source` | Where the code lives: `repo`, optional `ref` (a tag or commit), `path` inside the repository with forward slashes, and `region` |

A step where nothing goes over HTTP, such as a person approving in a browser, can have an example with samples and nothing else: the servers and the agent still run code for it.

### Code from Tested Files

Inline code drifts from the library it calls. A `source` points at a region of a real file instead, ideally a test that runs in that project's CI, so every sample compiles and runs with each release of the code it shows.

Mark the region with `pidl:begin` and `pidl:end` comments, in any comment syntax:

```go
func TestAgentIdentity(t *testing.T) {
	// ...setup...

	// pidl:begin agent-identity.request.agent
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resURL+"/api/documents", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := aauth.SignRequest(req, agent.Key, ""); err != nil {
		t.Fatal(err)
	}
	// pidl:end agent-identity.request.agent
}
```

The sample is the code between the markers. Marker lines of other regions inside it are dropped, the common indentation is removed, and blank lines at either end are trimmed. Naming regions after the step they show, `<scenario>.<step>.<role>`, keeps the two easy to match.

### The Lock File

Generating a walkthrough never reads the source repositories or the network. It reads a lock file beside the document, `<file without .json>.samples.json`, which holds each sample's code and its SHA-256:

```bash
# Read every source and write the lock file
pidl samples resolve aauth.json

# Fail when a source is not in the lock, its region is gone, or its code changed
pidl samples check aauth.json
```

`resolve` reads each repository from a checkout at `<root>/<repo>`, where the root is `--root`, or `$GOPATH/src` (`~/go/src`) by default. With a `ref`, it reads the file as it is at that tag or commit.

`generate -f walkthrough-json` and `generate -f viewer` read the lock by default; `--samples` names another one. Commit the lock with the document, and run `pidl samples check` in CI so samples cannot fall behind their sources. When the code they show changes on purpose, run `pidl samples resolve` again and commit the new lock.

### Narration for a Representation

A step's narration describes the protocol, but a reader looking at the Go tab may want to hear about the code. A step's `representations` override its narration for `http`, `curl` or a sample language. See [Narration](narration.md#representations).

## Related

- [Walkthrough Viewer](viewer.md)
- [Narration](narration.md)
- [Specification: Samples](../SPECIFICATION.md#samples)
- [CLI Reference: samples](../reference/cli.md#samples)
