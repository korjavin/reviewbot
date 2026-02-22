---
id: graphql-query-depth-complexity
title: GraphQL — Missing Query Depth and Complexity Limits (DoS)
severity: high
tags: [graphql, dos, depth-limit, complexity, resource-exhaustion]
taxonomy: security/graphql/dos
references:
  - https://cheatsheetseries.owasp.org/cheatsheets/GraphQL_Cheat_Sheet.html
  - https://www.howtographql.com/advanced/4-security/
---

# GraphQL — Missing Query Depth and Complexity Limits (DoS)

## Description

GraphQL allows clients to compose arbitrarily nested queries. Without depth and complexity limits, a single malicious query can trigger exponential database load — a "GraphQL DoS" or "query bomb". Deeply nested queries traverse relationships recursively, and high-complexity queries multiply resolver calls.

Classic attack: circular relationship abuse.
```graphql
{ user { friends { friends { friends { friends { friends { name } } } } } } }
```
If each `friends` call returns 100 users, depth=5 triggers 100^5 = 10 billion resolver calls.

## Vulnerable Pattern

```javascript
// BAD — Apollo Server with no depth or complexity limits
const server = new ApolloServer({
    typeDefs,
    resolvers,
    // No query depth limiting
    // No complexity analysis
    // Attacker can submit arbitrarily complex queries
});
```

```python
# BAD — graphene with no query limits
@app.route("/graphql", methods=["POST"])
def graphql_view():
    result = schema.execute(request.json["query"])
    return jsonify(result.data)
    # Single request can exhaust DB connections
```

## Secure Pattern

```javascript
// GOOD — Apollo Server with depth limiting and complexity analysis
const { ApolloServer } = require("@apollo/server");
const depthLimit = require("graphql-depth-limit");
const { createComplexityLimitRule } = require("graphql-validation-complexity");

const server = new ApolloServer({
    typeDefs,
    resolvers,
    validationRules: [
        depthLimit(7),  // max query depth of 7 levels
        createComplexityLimitRule(1000, {  // max complexity score of 1000
            scalarCost: 1,
            objectCost: 2,
            listFactor: 10,
            introspectionListFactor: 2,
        }),
    ],
});
```

```python
# GOOD — graphene with depth limiting
from graphql import parse, validate
from graphql.validation import NoSchemaIntrospectionCustomRule

MAX_DEPTH = 7

def check_query_depth(query_str: str) -> int:
    def get_depth(node, depth=0):
        if not hasattr(node, "selection_set") or not node.selection_set:
            return depth
        return max(get_depth(s, depth + 1) for s in node.selection_set.selections)
    ast = parse(query_str)
    return max(get_depth(d) for d in ast.definitions)

@app.route("/graphql", methods=["POST"])
def graphql_view():
    query = request.json.get("query", "")
    if check_query_depth(query) > MAX_DEPTH:
        return jsonify({"errors": [{"message": "Query depth exceeds limit"}]}), 400
    result = schema.execute(query)
    return jsonify(result.data)
```

## Checks to Generate

- Grep for GraphQL server setup missing `depthLimit` or `graphql-depth-limit` validation rule.
- Grep for `ApolloServer` without `validationRules` — no query validation constraints.
- Flag GraphQL endpoints with circular type relationships (`User → friends → User`) and no depth limit.
- Check for absence of query timeout: mutations and queries should have execution time limits.
- Grep for `graphql-validation-complexity` or equivalent — flag if missing on public-facing API.
- Flag GraphQL APIs allowing `@defer` or `@stream` without rate limiting (streaming amplification).
