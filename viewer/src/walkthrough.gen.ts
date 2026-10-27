// Code generated from schema/walkthrough.schema.json by scripts/types.mjs. DO NOT EDIT.
// Regenerate with: npm run types

/**
 * A protocol prepared for step-by-step presentation, produced by `pidl generate -f walkthrough-json`.
 */
export interface Walkthrough {
  /**
   * Format names the format and its major version, "pidl-walkthrough/v1".
   */
  format: "pidl-walkthrough/v1";
  id: string;
  name: string;
  description?: string;
  /**
   * Language is the source language of the narration.
   */
  language: string;
  pronunciations?: {
    [k: string]: {
      [k: string]: string;
    };
  };
  roles: OutlineRole[];
  groups?: Group[];
  scenarios: WalkthroughScenario[];
  tokens?: WalkthroughToken[];
  /**
   * Findings holds findings that do not locate a single flow, such as an
   * unused role.
   */
  findings?: Finding[];
}
/**
 * OutlineRole is one entity in an outline.
 *
 * This interface was referenced by `Walkthrough`'s JSON-Schema
 * via the `definition` "OutlineRole".
 */
export interface OutlineRole {
  position: number;
  id: string;
  name: string;
  type?: string;
  description?: string;
}
/**
 * Group is a navigation section that scenarios are listed under.
 *
 * This interface was referenced by `Walkthrough`'s JSON-Schema
 * via the `definition` "Group".
 */
export interface Group {
  /**
   * ID is the unique identifier scenarios refer to.
   */
  id: string;
  /**
   * Name is the display name.
   */
  name: string;
  /**
   * Description of the group.
   */
  description?: string;
}
/**
 * WalkthroughScenario is one scenario with its expanded steps.
 *
 * This interface was referenced by `Walkthrough`'s JSON-Schema
 * via the `definition` "WalkthroughScenario".
 */
export interface WalkthroughScenario {
  position: number;
  id: string;
  name: string;
  description?: string;
  fragment?: boolean;
  group?: string;
  variant_of?: string;
  variant_label?: string;
  narration?: Narration;
  briefing?: Briefing;
  samples?: Sample[];
  participants: string[];
  steps: WalkthroughStep[];
  /**
   * ObservedRun is set on a scenario entry that shows a run observed in a
   * real system: one instance of the scenario found in a capture.
   */
  observed_run?: ObservedRun;
  /**
   * ObservedExtra are the run's exchanges that matched no step, each
   * placed after the step it followed.
   */
  observed_extra?: ObservedExtra[];
}
/**
 * Narration is text written to be spoken, keyed by BCP 47 language tag.
 *
 * This interface was referenced by `Walkthrough`'s JSON-Schema
 * via the `definition` "Narration".
 */
export interface Narration {
  [k: string]: string;
}
/**
 * Briefing introduces a scenario before its steps, for example on slides that open a video.
 *
 * This interface was referenced by `Walkthrough`'s JSON-Schema
 * via the `definition` "Briefing".
 */
export interface Briefing {
  /**
   * Why explains, as prose, why the scenario exists and what problem it
   * solves.
   */
  why?: string;
  /**
   * Points are the key points, in display order.
   */
  points?: string[];
  /**
   * Narration is spoken over the briefing.
   */
  narration?: Narration;
}
/**
 * Sample is code that sends or handles a step's message, such as a Go SDK call.
 *
 * This interface was referenced by `Walkthrough`'s JSON-Schema
 * via the `definition` "Sample".
 */
export interface Sample {
  /**
   * Language names the code's language with a common highlighter name,
   * such as "go", "typescript", "python" or "csharp". Samples are also
   * representations a step's narration can be overridden for.
   */
  language: string;
  /**
   * Label distinguishes samples in one language, such as "Agent" and
   * "Resource".
   */
  label?: string;
  /**
   * Role is the entity ID whose developer writes this code.
   */
  role?: string;
  /**
   * Code is inline code, or a source's code once resolved. A sample has
   * Code, Source, or both.
   */
  code?: string;
  /**
   * Source refers to a marked region of a file in a repository.
   */
  source?: SampleSource;
}
/**
 * WalkthroughStep is one message of an expanded scenario.
 *
 * This interface was referenced by `Walkthrough`'s JSON-Schema
 * via the `definition` "WalkthroughStep".
 */
export interface WalkthroughStep {
  /**
   * Number is the 1-based position within the expanded scenario.
   */
  number: number;
  /**
   * ID is the step's flow ID, qualified as "<include>.<id>" for a step
   * inside an included fragment; empty when the flow has no ID.
   */
  id?: string;
  from: string;
  to: string;
  label: string;
  response?: boolean;
  phase?: string;
  /**
   * Fragments names the fragments this step sits inside, outermost first.
   */
  fragments?: string[];
  description?: string;
  narration?: Narration;
  /**
   * Representations holds narration overrides by representation.
   */
  representations?: {
    [k: string]: RepresentationNarration;
  };
  note?: string;
  signed_by?: string;
  issues?: string[];
  verifies?: string[];
  example?: FlowExample;
  /**
   * Observed is the exchange a real run produced for this step, in a
   * scenario entry with an ObservedRun. ObservedStatus is matched,
   * notSeen, skipped, unmatchable or unobservable; ObservedStrength is
   * strong or weak for a matched step.
   */
  observed?: ObservedMessage;
  observed_status?: string;
  observed_strength?: string;
  /**
   * Curl is a curl command equivalent to the example request.
   */
  curl?: string;
  location: string;
  findings?: Finding[];
}
/**
 * RepresentationNarration overrides a step's narration when the reader views the step in one representation, such as cURL or Go code.
 *
 * This interface was referenced by `Walkthrough`'s JSON-Schema
 * via the `definition` "RepresentationNarration".
 */
export interface RepresentationNarration {
  /**
   * Narration replaces the step's narration for this representation.
   */
  narration?: Narration;
}
/**
 * FlowExample is an illustrative message for one step: the HTTP request, the HTTP response and decoded tokens.
 *
 * This interface was referenced by `Walkthrough`'s JSON-Schema
 * via the `definition` "FlowExample".
 */
export interface FlowExample {
  /**
   * Source cites where the example comes from, such as "draft-11 §7.2.1".
   */
  source: string;
  /**
   * Kind is draft or derived.
   */
  kind: "draft" | "derived";
  /**
   * Request is the HTTP request, for steps that send one.
   */
  request?: ExampleRequest;
  /**
   * Response is the HTTP response, for steps that return one.
   */
  response?: ExampleResponse;
  /**
   * Tokens are decoded tokens the message carries, in display order.
   */
  tokens?: ExampleToken[];
  /**
   * Samples are code that sends or handles this message, such as SDK
   * calls, shown beside the HTTP form.
   */
  samples?: Sample[];
  /**
   * Note explains anything the reader should notice in this example.
   */
  note?: string;
}
/**
 * ExampleToken is a decoded token carried by a message.
 *
 * This interface was referenced by `Walkthrough`'s JSON-Schema
 * via the `definition` "ExampleToken".
 */
export interface ExampleToken {
  /**
   * Token is the token definition ID.
   */
  token: string;
  /**
   * Label says where the token travels, such as "Signature-Key header".
   */
  label?: string;
  /**
   * Header and Payload are the decoded JWT header and claims.
   */
  header?: unknown;
  payload?: unknown;
}
/**
 * Finding is an analysis result attached to a step.
 *
 * This interface was referenced by `Walkthrough`'s JSON-Schema
 * via the `definition` "Finding".
 */
export interface Finding {
  id: string;
  severity: string;
  title: string;
  description?: string;
  location?: string;
}
/**
 * ObservedExtra is an observed exchange that matched no step.
 *
 * This interface was referenced by `Walkthrough`'s JSON-Schema
 * via the `definition` "ObservedExtra".
 */
export interface ObservedExtra {
  /**
   * AfterStep is the number of the step it followed, or 0 when it came
   * before the first matched step.
   */
  after_step: number;
  /**
   * From and To are the participants, when the binding resolved them.
   */
  from?: string;
  to?: string;
  observed: ObservedMessage;
}
/**
 * ObservedMessage is an exchange observed in a real system, shown as a step's example is: request, response and decoded tokens.
 *
 * This interface was referenced by `Walkthrough`'s JSON-Schema
 * via the `definition` "ObservedMessage".
 */
export interface ObservedMessage {
  time?: string;
  /**
   * From and To are the capturing tool's labels for sender and receiver.
   */
  from?: string;
  to?: string;
  /**
   * RequestID is the value of a request ID header, when the capture has one.
   */
  request_id?: string;
  request?: ExampleRequest;
  response?: ExampleResponse;
  /**
   * Tokens are the tokens found in the message, decoded and not verified.
   */
  tokens?: ExampleToken[];
  /**
   * Label and Annotations are evidence the capturing tool recorded.
   */
  label?: string;
  annotations?: ExchangeAnnotation[];
  /**
   * Repeats is how many exchanges a repeated step took, when more than one.
   */
  repeats?: number;
  /**
   * BodyTruncated reports that a body was cut at the capture's size limit.
   */
  body_truncated?: boolean;
}
/**
 * ExampleRequest is an illustrative HTTP request.
 *
 * This interface was referenced by `Walkthrough`'s JSON-Schema
 * via the `definition` "ExampleRequest".
 */
export interface ExampleRequest {
  method: string;
  url: string;
  headers?: ExampleHeader[];
  /**
   * Body is the request body as JSON, kept byte for byte so key order
   * survives a round trip.
   */
  body?: unknown;
}
/**
 * ExampleHeader is one HTTP header.
 *
 * This interface was referenced by `Walkthrough`'s JSON-Schema
 * via the `definition` "ExampleHeader".
 */
export interface ExampleHeader {
  name: string;
  value: string;
}
/**
 * ExampleResponse is an illustrative HTTP response.
 *
 * This interface was referenced by `Walkthrough`'s JSON-Schema
 * via the `definition` "ExampleResponse".
 */
export interface ExampleResponse {
  status: number;
  headers?: ExampleHeader[];
  body?: unknown;
}
/**
 * ExchangeAnnotation is a piece of evidence about an exchange.
 *
 * This interface was referenced by `Walkthrough`'s JSON-Schema
 * via the `definition` "ExchangeAnnotation".
 */
export interface ExchangeAnnotation {
  /**
   * Kind classifies the annotation, such as "assertion" or "client-event".
   */
  kind: string;
  /**
   * Text is the annotation as a person reads it.
   */
  text: string;
  /**
   * OK is the result of an assertion; unset for other kinds.
   */
  ok?: boolean;
  /**
   * Time is when the annotated event happened, when it differs from the
   * exchange's time.
   */
  time?: string;
}
/**
 * WalkthroughToken is a token definition with the claims its checkers use.
 *
 * This interface was referenced by `Walkthrough`'s JSON-Schema
 * via the `definition` "WalkthroughToken".
 */
export interface WalkthroughToken {
  id: string;
  name: string;
  type?: string;
  issuer?: string;
  audience?: string;
  verified_by?: string[];
  claims?: TokenClaim[];
}
/**
 * TokenClaim is one claim in a token and the entities that check it.
 *
 * This interface was referenced by `Walkthrough`'s JSON-Schema
 * via the `definition` "TokenClaim".
 */
export interface TokenClaim {
  /**
   * Name is the claim name, such as "aud" or "agent_jkt".
   */
  name: string;
  /**
   * CheckedBy lists the entities that check this claim.
   */
  checked_by?: string[];
  /**
   * Description explains what the claim carries or how it is checked.
   */
  description?: string;
}
/**
 * ObservedRun summarizes one instance of a scenario found in a capture.
 *
 * This interface was referenced by `Walkthrough`'s JSON-Schema
 * via the `definition` "ObservedRun".
 */
export interface ObservedRun {
  /**
   * Scenario is the ID of the scenario this entry is an instance of.
   */
  scenario: string;
  /**
   * Source names the tool that captured the run; RequestFidelity says
   * whether its requests are the wire messages or a client's.
   */
  source: string;
  request_fidelity?: "wire" | "client";
  /**
   * Instance is this instance's number, from 1, of Instances found.
   */
  instance: number;
  instances: number;
  /**
   * From and To bound the instance's exchanges in time.
   */
  from?: string;
  to?: string;
  /**
   * RunnerUp is the next-best scenario at this point of the capture.
   */
  runner_up?: string;
  /**
   * Counts of step outcomes; Matched counts every matched step and Weak
   * the matched steps that are weak. Explained counts exchanges matched
   * to steps and Extra those that matched none.
   */
  matched: number;
  weak: number;
  not_seen: number;
  skipped: number;
  unmatchable: number;
  unobservable: number;
  explained: number;
  extra: number;
  /**
   * Redacted reports that raw token, signature and credential values
   * were masked.
   */
  redacted: boolean;
}
/**
 * SampleSource refers to the code between "pidl:begin <region>" and "pidl:end <region>" markers, in comments, in a file of a repository.
 *
 * This interface was referenced by `Walkthrough`'s JSON-Schema
 * via the `definition` "SampleSource".
 */
export interface SampleSource {
  /**
   * Repo is the repository's module-style path, such as
   * "github.com/aauth-dev/aauth-go".
   */
  repo: string;
  /**
   * Ref is a tag or commit to read the file at. Empty reads the
   * checkout's working tree.
   */
  ref?: string;
  /**
   * Path is the file's path within the repository.
   */
  path: string;
  /**
   * Region is the marked region's name.
   */
  region: string;
}
