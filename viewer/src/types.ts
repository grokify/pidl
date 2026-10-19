// Types for the walkthrough JSON produced by `pidl generate -f walkthrough-json`
// (pidl.Walkthrough in the Go library).

export interface Walkthrough {
  id: string;
  name: string;
  description?: string;
  roles: Role[];
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
  participants: string[];
  steps: Step[];
}

export interface Step {
  number: number;
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
