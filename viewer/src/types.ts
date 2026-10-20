// Types of the walkthrough format (pidl-walkthrough/v1), produced by
// `pidl generate -f walkthrough-json`. They are generated from the format's
// JSON Schema into walkthrough.gen.ts; this file gives them the names the
// viewer uses.

export type {
  Briefing,
  ExampleToken,
  Finding,
  Group,
  Narration,
  Sample,
  Walkthrough,
} from "./walkthrough.gen";

export type {
  TokenClaim as Claim,
  FlowExample as Example,
  ExampleHeader as Header,
  ExampleRequest as HttpRequest,
  ExampleResponse as HttpResponse,
  OutlineRole as Role,
  WalkthroughScenario as Scenario,
  WalkthroughStep as Step,
  WalkthroughToken as TokenDef,
} from "./walkthrough.gen";
