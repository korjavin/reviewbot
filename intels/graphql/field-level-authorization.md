---
id: graphql-field-level-authorization
title: GraphQL — Missing Field-Level Authorization
severity: critical
tags: [graphql, authorization, field-level, access-control, idor]
taxonomy: security/graphql/authorization
references:
  - https://cheatsheetseries.owasp.org/cheatsheets/GraphQL_Cheat_Sheet.html
  - https://www.graphql-shield.com/
---

# GraphQL — Missing Field-Level Authorization

## Description

GraphQL resolvers require authorization at every level — type, field, and argument. Unlike REST where endpoints map cleanly to permission checks, GraphQL lets clients request any combination of fields. A user querying `{ me { ... } }` can add `adminNotes`, `internalScore`, `paymentHistory` fields if those fields lack their own authorization checks.

Additionally, object-level auth (can user see this User?) must be combined with field-level auth (can user see `User.salary`?).

## Vulnerable Pattern

```javascript
// BAD — authorization only on the top-level query, not on sensitive fields
const resolvers = {
    Query: {
        user: async (_, { id }, { user }) => {
            if (!user) throw new AuthenticationError("Not authenticated");
            return db.users.findById(id);  // returns full User object
            // Fields like adminNotes, salary, passwordResetToken
            // are now accessible to ANY authenticated user
        }
    },
    User: {
        // No field-level resolvers — all fields returned to anyone who can see the User
        adminNotes: (user) => user.adminNotes,     // exposed!
        salary: (user) => user.salary,             // exposed!
        passwordResetToken: (user) => user.token,  // CRITICAL exposed!
    }
};
```

```graphql
# Attacker query — bypasses object-level auth to get sensitive fields:
query {
    user(id: 1) {
        name
        email
        adminNotes          # should be admin-only
        salary              # should be HR-only
        passwordResetToken  # should NEVER be exposed
    }
}
```

## Secure Pattern

```javascript
// GOOD — graphql-shield: declarative field-level permissions
const { shield, rule, and, or } = require("graphql-shield");

const isAuthenticated = rule({ cache: "contextual" })(
    async (parent, args, ctx) => ctx.user !== null
);
const isAdmin = rule({ cache: "contextual" })(
    async (parent, args, ctx) => ctx.user?.role === "admin"
);
const isSelf = rule({ cache: "no_cache" })(
    async (parent, args, ctx) => parent.id === ctx.user?.id
);

const permissions = shield({
    Query: {
        user: isAuthenticated,
    },
    User: {
        "*": isAuthenticated,          // default: must be authenticated
        adminNotes: isAdmin,           // only admins
        salary: and(isAdmin, isSelf),  // only admin seeing own salary (or HR role)
        passwordResetToken: deny,      // NEVER expose this field
        internalScore: isAdmin,
    },
});

const server = new ApolloServer({
    typeDefs,
    resolvers,
    plugins: [ApolloServerPluginShield(permissions)],
});
```

```javascript
// GOOD — alternatively: resolver-level auth guards
const resolvers = {
    User: {
        adminNotes: (user, _, { currentUser }) => {
            if (currentUser?.role !== "admin") return null;
            return user.adminNotes;
        },
        salary: (user, _, { currentUser }) => {
            if (currentUser?.role !== "hr" && currentUser?.id !== user.id) return null;
            return user.salary;
        },
    }
};
```

## Checks to Generate

- Flag GraphQL type resolvers without field-level permission checks on sensitive fields (`adminNotes`, `salary`, `internal*`, `*token*`, `*secret*`, `*key*`).
- Grep for resolver functions that return the full database object without field filtering.
- Check for absence of `graphql-shield` or equivalent permission framework.
- Flag User/Account type with `passwordResetToken`, `mfaSecret`, `apiKey` fields — should be excluded from schema or always null for non-owners.
- Grep for resolvers that only check `if (!user)` (authentication) but not role/ownership (authorization).
- Check `N+1` resolver patterns — field resolvers called per-object without DataLoader may also bypass batched auth checks.
