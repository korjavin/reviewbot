---
id: graphql-introspection-exposure
title: GraphQL — Introspection Enabled in Production
severity: medium
tags: [graphql, introspection, api, information-disclosure, reconnaissance]
taxonomy: security/graphql/introspection
references:
  - https://owasp.org/www-project-web-security-testing-guide/
  - https://cheatsheetseries.owasp.org/cheatsheets/GraphQL_Cheat_Sheet.html
  - https://graphql.org/learn/introspection/
---

# GraphQL — Introspection Enabled in Production

## Description

GraphQL introspection allows any client to query the complete API schema — all types, fields, queries, mutations, arguments, and their descriptions. Enabled in production, it gives attackers a full map of the API surface including internal fields, admin mutations, and deprecated endpoints that may lack proper authorization checks.

Introspection queries are the first step in GraphQL reconnaissance:
```graphql
{ __schema { types { name fields { name } } } }
```

## Vulnerable Pattern

```python
# BAD — graphene (Python): introspection enabled by default
from graphene import Schema
schema = Schema(query=Query, mutation=Mutation)
# Introspection is on by default — no configuration to disable it

# Flask + graphene: no introspection restriction
@app.route("/graphql", methods=["POST"])
def graphql_endpoint():
    data = request.get_json()
    result = schema.execute(data["query"])
    return jsonify(result.data)
```

```javascript
// BAD — Apollo Server: introspection on in all environments
const server = new ApolloServer({
    typeDefs,
    resolvers,
    // introspection defaults to true in all environments
});

// BAD — explicitly enabled in production
const server = new ApolloServer({
    typeDefs,
    resolvers,
    introspection: true,  // hardcoded on — ignores NODE_ENV
});
```

## Secure Pattern

```javascript
// GOOD — Apollo Server: disable introspection in production
const server = new ApolloServer({
    typeDefs,
    resolvers,
    introspection: process.env.NODE_ENV !== "production",
    // Also disable playground in production
    playground: process.env.NODE_ENV !== "production",
});
```

```python
# GOOD — graphene with introspection blocked via middleware
from graphql import GraphQLError

class DisableIntrospectionMiddleware:
    def resolve(self, next, root, info, **kwargs):
        if info.field_name.startswith("__"):
            raise GraphQLError("Introspection is disabled")
        return next(root, info, **kwargs)

result = schema.execute(
    query,
    middleware=[DisableIntrospectionMiddleware()],
)
```

```python
# GOOD — block introspection at query parsing level
INTROSPECTION_QUERY_PATTERN = re.compile(r"__schema|__type|__typename", re.IGNORECASE)

@app.route("/graphql", methods=["POST"])
def graphql_endpoint():
    data = request.get_json()
    query = data.get("query", "")
    if INTROSPECTION_QUERY_PATTERN.search(query) and not current_user.is_admin:
        return jsonify({"errors": [{"message": "Introspection disabled"}]}), 403
    result = schema.execute(query)
    return jsonify(result.data)
```

## Checks to Generate

- Grep for `introspection: true` in Apollo Server config — flag if not conditioned on non-production env.
- Grep for GraphQL schema setup without introspection-disabling middleware.
- Check GraphQL endpoint responds to `{ __schema { types { name } } }` — automated probe.
- Flag GraphQL playground/GraphiQL UI enabled in production (`playground: true`, `/graphql` GET returns UI).
- Grep for missing `NODE_ENV` check around GraphQL introspection configuration.
