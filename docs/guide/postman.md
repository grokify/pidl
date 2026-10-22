# Postman Collections

`pidl generate -f postman` turns a PIDL file into a Postman collection you can run: a folder for each scenario, a request for each step the client role sends, values passed from each response to the next request, and checks on every response. The collection is generated from the same examples the walkthrough viewer shows, so the two agree step for step.

```bash
# Signed requests sent through a signing proxy
pidl generate -f postman --signer proxy -o aauth.postman_collection.json aauth_protocol

# Requests sent directly, with saved example responses for a Postman mock server
pidl generate -f postman --mock -o aauth-mock.postman_collection.json aauth_protocol
```

## What Is Generated

| Part | From |
|------|------|
| Folders | Groups, then scenarios, in output order; fragments expanded in place |
| Requests | Steps whose sender is the client role (`--client`, default: the first `agent` or `client` entity) and that have an example request |
| URLs | Each role's base URL becomes a variable such as `{{ps_url}}` |
| Chained values | A value an earlier response delivers (a body member, a `Location` or another header, or a string parameter of a structured header such as the `resource-token` in `AAuth-Requirement`) and a later request reuses is saved to a collection variable and referenced as `{{name}}` |
| Checks | Status code; headers named in the example; body members; and the claim names of tokens the body carries |
| Deferred responses | A request answered `202` with `Location` accepts either the `202` (then the next request polls `{{location}}` until it completes) or the final response (then the poll is skipped) |
| Folder notes | The scenario's description and narration, and the steps between other roles that the collection does not send |

## Signing Proxies

Postman cannot create HTTP message signatures ([RFC 9421](https://www.rfc-editor.org/rfc/rfc9421)). With `--signer proxy`, each request whose example is signed goes to `{{signer_url}}/<target URL>`, and a signing proxy signs it and forwards it. Requests without signatures are sent directly.

| Header | Sent to the proxy |
|--------|-------------------|
| `Signature`, `Signature-Input`, `Content-Digest` | Left out; the proxy creates them |
| `Signature-Key` | How to sign: the example's label and scheme, such as `sig=hwk`. For the `jwt` scheme it also carries the token to present when an earlier response delivered it, `sig=jwt;jwt="{{user_token}}"`; without one, the proxy presents its own |
| Everything else | As in the example, with chained values |

The proxy decides what each token means for its protocol. The AAuth sandbox in [aauth-go](https://github.com/aauth-dev/aauth-go) is such a proxy for AAuth.

## Environments

An environment says where a target's roles run. Write targets in a JSON file:

```json
{
  "sandbox": {
    "signer": "http://127.0.0.1:8099",
    "roles": {
      "ap": "https://ap.aauth.test",
      "ps": "https://ps.aauth.test",
      "resource": "https://resource.aauth.test",
      "as": "https://as.aauth.test"
    },
    "scenarios": {
      "federated-authorization": {"resource": "https://docs.aauth.test"}
    }
  },
  "mock": {
    "roles": {"ap": "{{mock_url}}", "ps": "{{mock_url}}", "resource": "{{mock_url}}", "as": "{{mock_url}}"},
    "variables": {"mock_url": "https://<your-mock-id>.mock.pstmn.io"}
  }
}
```

```bash
pidl generate -f postman-environment --targets targets.json --target sandbox -o sandbox.postman_environment.json aauth_protocol
```

`scenarios` points a role at another server for one scenario's folder only: here, federated authorization uses a resource that has its own access server. Without `--targets`, the environment uses the base URLs of the examples.

## Running

Import the collection and an environment into Postman and run a folder, or use Newman:

```bash
npx newman run aauth.postman_collection.json -e sandbox.postman_environment.json --folder "Federated authorization (four-party)"
```

Collections are generated output: edit the PIDL file and regenerate rather than editing the collection.
