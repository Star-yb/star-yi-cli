# Sa-Token 动态角色/权限注入说明（RBAC + 缓存）

本文档用于解释本项目中 **Sa-Token 的角色与权限动态注入逻辑**，以及当用户角色/权限发生变更时，系统应该如何处理缓存与会话，确保鉴权结果及时生效。

对应实现代码：

- `yi-admin/src/main/java/com/star/admin/common/satoken/StpInterfaceImpl.java`

---

## 1. 背景：Sa-Token 如何获取“角色/权限”

Sa-Token 在执行鉴权时（例如 `StpUtil.checkRole(...)` / `StpUtil.checkPermission(...)` / 注解鉴权），会回调 `StpInterface`：

- `getRoleList(loginId, loginType)`：返回当前账号拥有的 **角色标识集合**
- `getPermissionList(loginId, loginType)`：返回当前账号拥有的 **权限码集合**

因此，只要在 `StpInterfaceImpl` 中实现好这两个方法，Sa-Token 就能做到“动态注入”（按需加载）角色和权限。

---

## 2. 本项目的 RBAC 关系模型（基于实体关联）

结合目前项目的实体关联（Jimmer Table 反查/关联）：

- **用户（Users）** ↔ **角色（Roles）**：多对多
- **角色（Roles）** ↔ **权限（Permissions）**：多对多

鉴权时的推导关系：

1. 先根据 `loginId` 查询该用户的 **角色 roleCode 列表**
2. 再根据每个 `roleCode` 查询该角色的 **权限 permissionCode 列表**
3. 将多个角色的权限合并（去重由查询 `distinct()` 或上层自行处理）

---

## 3. 数据查询方式（使用 Jimmer Table，与你的 TestController 风格一致）

### 3.1 用户 → 角色（getRoleList）

当前实现使用 **从 Roles 反查 Users** 的写法（与 `TestController` 一致）：

- 入口表：`RolesTable`
- 条件：`r.users(u -> u.id().eq(userId))`
- 选择：`r.roleCode()`
- 去重：`distinct()`

得到：`List<String> roleCodeList`

### 3.2 角色 → 权限（getPermissionList）

当前实现按角色逐个查询权限：

- 入口表：`PermissionsTable`
- 条件：`p.roles(r -> r.roleCode().eq(roleCode))`
- 选择：`p.permissionCode()`
- 去重：`distinct()`

得到：`List<String> permissionCodeList`

---

## 4. 缓存策略（SaTokenDao）

为了减少每次鉴权都打数据库，本项目按官方推荐使用 `SaManager.getSaTokenDao()` 做缓存。

### 4.1 缓存 Key 约定

- **用户 → 角色缓存**
  - Key：`satoken:loginId-find-role:{loginId}`
  - Value：`List<String>`（roleCode 列表）
  - TTL：30 天（当前实现为 `60 * 60 * 24 * 30` 秒）

- **角色 → 权限缓存（按角色维度缓存）**
  - Key：`satoken:role-find-permission:{roleCode}`
  - Value：`List<String>`（permissionCode 列表）
  - TTL：30 天（同上）

### 4.2 读取优先级

- **先读缓存**（`getObject(key)`）
- 缓存 miss：**再查数据库**
- 查到后：`setObject(key, value, ttlSeconds)` 写回缓存

这样做到：

- 鉴权高频时，大部分请求只命中缓存
- 只有首次或缓存过期时才访问数据库

---

## 5. “动态”的含义与边界

这里的“动态”不是指“每次请求都实时查库”，而是：

- Sa-Token 在鉴权时 **动态回调** 你的 `StpInterfaceImpl`
- `StpInterfaceImpl` 内部 **按需加载**（缓存 miss 才查库）

因此，鉴权的实时性取决于：**缓存是否被及时刷新/删除**。

---

## 6. 当用户的角色/权限发生变动时，应该怎么做？

本项目的缓存分两层（用户→角色、角色→权限），所以变更时要按影响范围处理。

### 6.1 变更场景与影响

- **场景 A：给某个用户新增/移除角色**
  - 影响：该用户的 `roleCodeList` 变化
  - 需要处理的缓存：`satoken:loginId-find-role:{userId}`
  - 间接影响：该用户的权限集合会随角色变化而变化（但权限缓存按角色维度，不一定需要动）

- **场景 B：给某个角色新增/移除权限**
  - 影响：该角色对应的 `permissionCodeList` 变化
  - 需要处理的缓存：`satoken:role-find-permission:{roleCode}`
  - 间接影响：所有拥有该角色的用户，权限集合都会变化（但用户→角色缓存不一定需要动）

- **场景 C：修改角色编码 roleCode**
  - 影响：缓存 key 本身会变化（旧 key 失效，新 key 需要重新写入）
  - 建议：尽量避免“修改 roleCode”，更多使用“角色名称 roleName”展示，roleCode 作为稳定标识

### 6.2 推荐做法：变更后删除缓存 key（让下一次鉴权自动回源）

这是最简单稳定的做法：**不强行更新缓存内容，只删除**，让下次鉴权自动重新查询并写入。

- 用户角色变更后：
  - `SaManager.getSaTokenDao().deleteObject("satoken:loginId-find-role:" + userId);`

- 角色权限变更后：
  - `SaManager.getSaTokenDao().deleteObject("satoken:role-find-permission:" + roleCode);`

这样做的效果：

- 用户下一次鉴权时，会立刻从数据库重新计算角色/权限
- 不需要关心旧缓存 TTL 是否很长

### 6.3 是否需要“踢下线/刷新会话”？

取决于你希望变更生效的时机：

- **希望“下次请求”就生效**：
  - 仅删除缓存 key 一般就够了（因为下一次鉴权会重新回源）

- **希望“立即生效且强制”**（例如立刻收回权限）：
  - 除了删缓存，还可以选择：
    - **踢下线**：让用户重新登录获取新的权限环境
    - 或 **刷新/重建会话数据**（如果你把某些权限信息缓存到了 Session）

> 当前项目的角色/权限来源以 `StpInterfaceImpl + SaTokenDao` 为主，并未强依赖 Session 存角色/权限，因此“删缓存”通常足够。

---

## 7. 你可以在项目里做哪些增强（可选）

### 7.1 提供统一的“RBAC 缓存清理工具类”

建议在 `yi-common` 或 `yi-admin` 增加一个小工具（例如 `RbacCacheHelper`）封装 key 与清理逻辑，避免散落在各处写字符串 key：

- `clearUserRolesCache(userId)`
- `clearRolePermissionsCache(roleCode)`

### 7.2 在“分配角色/分配权限”的 Service 方法里自动清理缓存

当你实现：

- 用户分配角色：`assignRoles(userId, roleIds)` / `removeRoles(userId, roleIds)`
- 角色分配权限：`assignPermissions(roleId, permissionIds)` / `removePermissions(roleId, permissionIds)`

在数据库写入成功后，立刻删除相关缓存 key，即可保证鉴权动态刷新。

### 7.3 可观察性：统计缓存命中率/回源次数

如果你后期想优化性能，可以在 `StpInterfaceImpl` 中（用日志框架）记录：

- roleList 缓存命中/回源次数
- permissionList 缓存命中/回源次数

用于评估是否需要改 TTL、是否需要批量查询等。

---

## 8. 小结

- **角色来源**：`RolesTable` 反查 `users(id)` 得到 `roleCodeList`
- **权限来源**：`PermissionsTable` 通过 `roles(roleCode)` 得到 `permissionCodeList`
- **缓存分层**：
  - 用户→角色：`satoken:loginId-find-role:{loginId}`
  - 角色→权限：`satoken:role-find-permission:{roleCode}`
- **变更后推荐动作**：删除对应缓存 key，让下一次鉴权自动回源刷新  
  - 用户角色变更：删用户→角色缓存
  - 角色权限变更：删角色→权限缓存

