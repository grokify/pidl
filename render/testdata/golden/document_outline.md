# Tokens Demo

## Roles

1. **Agent**
2. **Resource**
3. **Person Server**
4. **Access Server**

## Scenarios

### 1. Federated authorization

Participants: Agent, Resource, Person Server, Access Server

1. Agent → Resource: request
2. Resource ⇢ Agent: challenge
3. Agent → Person Server: token_request
4. See *PS to AS* (scenario 2)
5. Agent → Resource: retry

### 2. PS to AS (fragment)

Participants: Person Server, Access Server

1. Person Server → Access Server: exchange
2. Access Server ⇢ Person Server: auth_token

## Tokens

- **Resource token**: issued by Resource; for Person Server; checked by Person Server, Access Server
- **Auth token**: issued by Access Server; for Resource; checked by Resource
- **Session token**
