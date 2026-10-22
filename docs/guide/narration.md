# Narration

Narration is text written to be spoken: the script for audio, autoplay and narrated video of a walkthrough. It is separate from `description` because on-screen prose reads badly aloud. A description can say "the server compares the SHA-256 of `code_verifier` with `code_challenge`"; a narrator should say "the server checks the secret against the challenge".

PIDL keeps narration in the document, beside the steps it describes, and passes it into the [walkthrough](../reference/walkthrough-format.md), where audio and video tools read it. The walkthrough viewer does not play narration yet.

## Narration by Language

Every narration value is an object keyed by [BCP 47](https://www.rfc-editor.org/info/bcp47) language tag. Scenarios and flows carry it:

```json
{
  "protocol": {"id": "oauth2-api-access", "name": "OAuth 2.0 API Access", "language": "en-US"},
  "scenarios": [{
    "id": "authorization-code",
    "name": "Authorization code with PKCE",
    "narration": {
      "en-US": "A user lets an application read their reports.",
      "fr-FR": "Un utilisateur autorise une application à lire ses rapports."
    },
    "flows": [{
      "id": "token-request",
      "from": "client", "to": "as", "action": "token_request",
      "narration": {"en-US": "The application trades the code for a token."}
    }]
  }]
}
```

`protocol.language` is the language the document is written in, `en-US` unless set. Every narration object must include it, and it is the fallback for any language a narration lacks, so a partly translated document still narrates every step.

Include steps carry no narration; the steps of their fragment do.

## Briefings

A briefing introduces a scenario before its first step: why it matters, and what to watch for. Video tools show it as an opening slide.

```json
"briefing": {
  "why": "The client never sees the user's credentials, and a stolen code is useless without the PKCE verifier.",
  "points": ["The browser carries the code, not the token", "PKCE binds the code to the client that asked for it"],
  "narration": {"en-US": "Why the authorization code grant with PKCE is the default."}
}
```

## Representations

A reader can view a step as HTTP, as a `curl` command or as a [code sample](samples.md). A step's `representations` override its narration for one of those views, so the Go tab can talk about the code rather than the wire format:

```json
{
  "id": "token-request",
  "from": "client", "to": "as", "action": "token_request",
  "narration": {"en-US": "The application trades the code for a token."},
  "representations": {
    "curl": {"narration": {"en-US": "This curl command posts the form the token endpoint expects."}},
    "go": {"narration": {"en-US": "Exchange sends the code and the verifier in one call."}}
  }
}
```

The keys are `http`, `curl`, or a language of one of the step's samples. For a reader's language and representation, the spoken text is the first of these that exists:

1. The representation's narration in the reader's language
2. The representation's narration in the source language
3. The step's narration in the reader's language
4. The step's narration in the source language
5. The step's `description`

## Pronunciations

Pronunciations map a term to how it is spoken, per language:

```json
"pronunciations": {
  "PKCE": {"en-US": "pixy"},
  "OAuth": {"en-US": "oh auth"}
}
```

They use the shape [videoascode](https://github.com/grokify/videoascode) reads, so text-to-speech tools can use them unchanged. Pronunciations from documents named by `extends` and `imports` are merged in.

## Validation

`pidl validate` checks that narration keys are BCP 47 tags with non-empty text, that every narration object includes the source language, that representation keys are `http`, `curl` or a sample language on that step, and that include steps carry no narration.

## Related

- [Scenarios](scenarios.md)
- [Code Samples and cURL](samples.md)
- [Specification: Narration](../SPECIFICATION.md#narration)
