---
id: graphql-batching-dos
title: GraphQL — Query Batching and Alias Abuse for DoS / Brute Force
severity: high
tags: [graphql, batching, alias, brute-force, dos, rate-limiting]
taxonomy: security/graphql/batching
references:
  - https://cheatsheetseries.owasp.org/cheatsheets/GraphQL_Cheat_Sheet.html
  - https://lab.wallarm.com/graphql-batching-attack/
---

# GraphQL — Query Batching and Alias Abuse for DoS / Brute Force

## Description

GraphQL supports two mechanisms attackers abuse to bypass rate limiting:

1. **Query Batching**: Sending an array of operations in a single HTTP request. Rate limiting per-request fails because 1000 login attempts fit in 1 request.
2. **Alias Abuse**: Using GraphQL aliases to repeat the same field/mutation many times in one query. Each alias invokes the resolver separately.

These techniques allow brute-forcing passwords, OTPs, and API keys through a single HTTP request that looks like one operation to rate limiters.

## Vulnerable Pattern

```graphql
# BAD — alias abuse: 1000 login attempts in 1 HTTP request
mutation {
  login1: login(username: "admin", password: "password1") { token }
  login2: login(username: "admin", password: "password2") { token }
  login3: login(username: "admin", password: "password3") { token }
  # ... login1000: ...
}
```

```json
// BAD — batched operations: 500 OTP checks in 1 request
[
  {"query": "mutation { verifyOTP(code: \"000000\") { success } }"},
  {"query": "mutation { verifyOTP(code: \"000001\") { success } }"},
  ...
]
```

```javascript
// BAD — Apollo Server with batching enabled (default) and no per-operation rate limiting
const server = new ApolloServer({
    typeDefs,
    resolvers,
    // allowBatchedHttpRequests defaults to true in older versions
});
```

## Secure Pattern

```javascript
// GOOD — disable batching OR limit batch size
const server = new ApolloServer({
    typeDefs,
    resolvers,
    allowBatchedHttpRequests: false,  // disable batching entirely
    // OR: limit batch size in middleware
});

// GOOD — custom validation rule to limit alias repetition
const MaxAliasesRule = (maxAliases) => (context) => ({
    Field(node) {
        // count total fields including aliases across the operation
    },
    OperationDefinition(node) {
        const aliases = countAliases(node);
        if (aliases > maxAliases) {
            context.reportError(new GraphQLError(`Max ${maxAliases} aliases allowed`));
        }
    }
});

const server = new ApolloServer({
    validationRules: [MaxAliasesRule(10)],
});
```

```javascript
// GOOD — rate limit at the operation level, not just HTTP request level
const rateLimiter = new RateLimiter({ points: 10, duration: 60 });

app.use("/graphql", async (req, res, next) => {
    const operations = Array.isArray(req.body) ? req.body : [req.body];
    // Charge rate limit per operation, not per request
    try {
        await rateLimiter.consume(req.ip, operations.length);
        next();
    } catch {
        res.status(429).json({ error: "Too many requests" });
    }
});
```

## Checks to Generate

- Grep for `allowBatchedHttpRequests: true` or absence of batching config in Apollo Server.
- Flag GraphQL APIs on sensitive mutations (login, OTP, password reset) without alias count limits.
- Grep for rate-limiting middleware applied at HTTP level only — must count GraphQL operations.
- Flag missing `graphql-rate-limit` or equivalent per-field/per-resolver rate limiting.
- Check login, verifyOTP, verifyEmail, resetPassword resolvers for per-resolver rate limiting.
- Flag batch size with no upper bound — even if batching is allowed, cap at 10–20 operations.
