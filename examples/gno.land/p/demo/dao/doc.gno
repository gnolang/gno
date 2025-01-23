// Package dao houses common DAO building blocks (framework), which can be used or adopted by any
// specific DAO implementation. By design, the DAO should house the proposals it receives, but not the actual
// DAO members or proposal votes. These abstractions should be implemented by a separate entity, to keep the DAO
// agnostic of implementation details such as these (member / vote management).
package dao
