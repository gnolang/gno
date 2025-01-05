# Allow read, write, and delete operations only on the user's own path
path "kv/data/{{identity.entity.name}}/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

# Deny access to all other KV paths
path "kv/*" {
  capabilities = []
}