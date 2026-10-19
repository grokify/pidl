// Types for the walkthrough JSON produced by `pidl generate -f walkthrough-json`
// (pidl.Walkthrough in the Go library).

export interface Walkthrough {
  id: string;
  name: string;
  description?: string;
  // language is the narration source language.
  language?: string;
  pronunciations?: Record<string, Record<string, string>>;
  roles: Role[];
  groups?: Group[];
  scenarios: Scenario[];
  tokens?: TokenDef[];
  findings?: Finding[];
}

export interface Role {
  position: number;
  id: string;
  name: string;
  type?: string;
  description?: string;
}

export interface Scenario {
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
  participants: string[];
  steps: Step[];
}

// Narration is spoken text keyed by BCP 47 language tag.
export type Narration = Record<string, string>;

export interface Briefing {
  why?: string;
  points?: string[];
  narration?: Narration;
}

export interface Group {
  id: string;
  name: string;
  description?: string;
}

export interface Step {
  number: number;
  // id is the step's flow ID, "<include>.<id>" inside a fragment.
  id?: string;
  narration?: Narration;
  representations?: Record<string, { narration?: Narration }>;
  // curl is a curl command equivalent to the example request.
  curl?: string;
  from: string;
  to: string;
  label: string;
  response?: boolean;
  phase?: string;
  fragments?: string[];
  description?: string;
  note?: string;
  signed_by?: string;
  issues?: string[];
  verifies?: string[];
  example?: Example;
  location: string;
  findings?: Finding[];
}

export interface Example {
  source: string;
  kind: "draft" | "derived";
  request?: HttpRequest;
  response?: HttpResponse;
  tokens?: ExampleToken[];
  note?: string;
}

export interface Header {
  name: string;
  value: string;
}

export interface HttpRequest {
  method: string;
  url: string;
  headers?: Header[];
  body?: unknown;
}

export interface HttpResponse {
  status: number;
  headers?: Header[];
  body?: unknown;
}

export interface ExampleToken {
  token: string;
  label?: string;
  header?: Record<string, unknown>;
  payload?: Record<string, unknown>;
}

export interface TokenDef {
  id: string;
  name: string;
  type?: string;
  issuer?: string;
  audience?: string;
  verified_by?: string[];
  claims?: Claim[];
}

export interface Claim {
  name: string;
  checked_by?: string[];
  description?: string;
}

export interface Finding {
  id: string;
  severity: string;
  title: string;
  description?: string;
  location?: string;
}
