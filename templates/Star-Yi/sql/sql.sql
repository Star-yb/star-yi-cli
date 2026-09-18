CREATE TABLE `sys_users` (
                             `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
                             `username` VARCHAR(50) NOT NULL COMMENT '用户名',
                             `password` VARCHAR(100) NOT NULL COMMENT '密码(加密)',
                             `nickname` VARCHAR(50) DEFAULT NULL COMMENT '昵称/真实姓名',
                             `email` VARCHAR(100) DEFAULT NULL COMMENT '邮箱',
                             `phone` VARCHAR(20) DEFAULT NULL COMMENT '手机号',
                             `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0:正常 1:禁用 2:删除',
                             `is_super_admin` TINYINT NOT NULL DEFAULT 0 COMMENT '是否超级管理员 0:否 1:是',
                             `last_login_time` DATETIME DEFAULT NULL COMMENT '最后登录时间',
                             `created_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                             `modified_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
                             `deleted_time` DATETIME DEFAULT NULL COMMENT '删除时间',
                             `is_deleted` TINYINT NOT NULL DEFAULT 0 COMMENT '是否删除 0:否 1:是',
                             PRIMARY KEY (`id`),
                             UNIQUE KEY `uk_username` (`username`),
                             UNIQUE KEY `uk_email` (`email`),
                             UNIQUE KEY `uk_phone` (`phone`),
                             KEY `idx_status` (`status`),
                             KEY `idx_deleted` (`is_deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';


CREATE TABLE `sys_roles` (
                             `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
                             `role_code` VARCHAR(50) NOT NULL COMMENT '角色编码(如: ADMIN, USER)',
                             `role_name` VARCHAR(50) NOT NULL COMMENT '角色名称(如: 管理员, 普通用户)',
                             `description` VARCHAR(200) DEFAULT NULL COMMENT '角色描述',
                             `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0:正常 1:禁用 2:删除',
                             `created_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                             `modified_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
                             `deleted_time` DATETIME DEFAULT NULL COMMENT '删除时间',
                             `is_deleted` TINYINT NOT NULL DEFAULT 0 COMMENT '是否删除 0:否 1:是',
                             PRIMARY KEY (`id`),
                             UNIQUE KEY `uk_role_code` (`role_code`),
                             KEY `idx_status` (`status`),
                             KEY `idx_deleted` (`is_deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色表';


CREATE TABLE `sys_permissions` (
                                   `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
                                   `permission_code` VARCHAR(100) NOT NULL COMMENT '权限编码(如: user:add, user:delete)',
                                   `permission_name` VARCHAR(50) NOT NULL COMMENT '权限名称(如: 用户新增, 用户删除)',
                                   `permission_type` TINYINT NOT NULL DEFAULT 1 COMMENT '权限类型 1:菜单 2:按钮 3:接口',
                                   `parent_id` BIGINT DEFAULT 0 COMMENT '父级权限ID(0表示顶级)',
                                   `path` VARCHAR(200) DEFAULT NULL COMMENT '前端路由路径',
                                   `icon` VARCHAR(50) DEFAULT NULL COMMENT '图标',
                                   `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序号',
                                   `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0:正常 1:禁用 2:删除',
                                   `created_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                   `modified_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
                                   `deleted_time` DATETIME DEFAULT NULL COMMENT '删除时间',
                                   `is_deleted` TINYINT NOT NULL DEFAULT 0 COMMENT '是否删除 0:否 1:是',
                                   PRIMARY KEY (`id`),
                                   UNIQUE KEY `uk_permission_code` (`permission_code`),
                                   KEY `idx_parent_id` (`parent_id`),
                                   KEY `idx_status` (`status`),
                                   KEY `idx_deleted` (`is_deleted`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='权限表';


CREATE TABLE `sys_user_role` (
                                 `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
                                 `user_id` BIGINT NOT NULL COMMENT '用户 ID',
                                 `role_id` BIGINT NOT NULL COMMENT '角色 ID',
                                 `created_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                 `created_by` BIGINT DEFAULT NULL COMMENT '创建人 ID(可为空)',
                                 PRIMARY KEY (`id`),
                                 UNIQUE KEY `uk_user_role` (`user_id`, `role_id`),
                                 KEY `idx_user_id` (`user_id`),
                                 KEY `idx_role_id` (`role_id`),
                                 CONSTRAINT `fk_user_role_user` FOREIGN KEY (`user_id`) REFERENCES `sys_users` (`id`) ON DELETE CASCADE,
                                 CONSTRAINT `fk_user_role_role` FOREIGN KEY (`role_id`) REFERENCES `sys_roles` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户 - 角色关联表';


CREATE TABLE `sys_role_permission` (
                                       `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键 ID',
                                       `role_id` BIGINT NOT NULL COMMENT '角色 ID',
                                       `permission_id` BIGINT NOT NULL COMMENT '权限 ID',
                                       `created_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
                                       `created_by` BIGINT DEFAULT NULL COMMENT '创建人 ID(可为空)',
                                       PRIMARY KEY (`id`),
                                       UNIQUE KEY `uk_role_permission` (`role_id`, `permission_id`),
                                       KEY `idx_role_id` (`role_id`),
                                       KEY `idx_permission_id` (`permission_id`),
                                       CONSTRAINT `fk_role_permission_role` FOREIGN KEY (`role_id`) REFERENCES `sys_roles` (`id`) ON DELETE CASCADE,
                                       CONSTRAINT `fk_role_permission_permission` FOREIGN KEY (`permission_id`) REFERENCES `sys_permissions` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色 - 权限关联表';


CREATE TABLE `sys_login_log` (
                                 `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
                                 `user_id` BIGINT DEFAULT NULL COMMENT '用户ID(逻辑关联，无外键；删除用户后仍保留原值)',
                                 `username` VARCHAR(50) NOT NULL COMMENT '登录账号(用户名/手机号)',
                                 `login_type` TINYINT NOT NULL DEFAULT 1 COMMENT '操作类型 1:登录 2:登出',
                                 `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态 0:成功 1:失败',
                                 `fail_reason` VARCHAR(200) DEFAULT NULL COMMENT '失败原因',
                                 `ip_address` VARCHAR(50) DEFAULT NULL COMMENT '登录IP',
                                 `login_location` VARCHAR(200) DEFAULT NULL COMMENT '登录地点',
                                 `browser` VARCHAR(100) DEFAULT NULL COMMENT '浏览器',
                                 `os` VARCHAR(100) DEFAULT NULL COMMENT '操作系统',
                                 `user_agent` VARCHAR(500) DEFAULT NULL COMMENT 'User-Agent',
                                 `token_value` VARCHAR(64) DEFAULT NULL COMMENT 'Token标识(脱敏,便于会话追踪)',
                                 `login_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录/登出时间',
                                 PRIMARY KEY (`id`),
                                 KEY `idx_user_id` (`user_id`),
                                 KEY `idx_username` (`username`),
                                 KEY `idx_login_time` (`login_time`),
                                 KEY `idx_status` (`status`),
                                 KEY `idx_login_type` (`login_type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='登录日志表';

-- ==========================================================
-- 模拟数据
-- ==========================================================
-- 用法：
--   1. 【推荐】直接执行：仅导入 超级管理员 + 基础角色/权限
--   2. 【可选】如需测试用户(user1/user2/user3)，三选一：
--        a. 编辑器中取消 [2/2] 整段注释（删除行首 -- 标记）
--        b. MySQL 客户端手动选中 [2/2] 段执行
--        c. 命令行提取本段：
--           sed -n '/-- \[2\/2\] 可选数据 START/,/-- \[2\/2\] 可选数据 END/p' sql.sql
-- ==========================================================

-- ==========================================================
-- [1/2] 必选数据：超级管理员 + 基础角色/权限
-- 说明：执行本段即可获得可登录的管理后台
--       默认账号: admin / 123456 (SHA-256 哈希存储)
-- ==========================================================

-- 1.1 角色数据
INSERT INTO `sys_roles` (`role_code`, `role_name`, `description`, `status`) VALUES
('SUPER_ADMIN', '超级管理员', '拥有系统所有权限', 0),
('ADMIN', '管理员', '拥有系统管理权限', 0),
('USER', '普通用户', '普通用户权限', 0);

-- 1.2 权限数据
INSERT INTO `sys_permissions` (`permission_code`, `permission_name`, `permission_type`, `parent_id`, `path`, `icon`, `sort_order`, `status`) VALUES
-- 顶级菜单
('system', '系统管理', 1, 0, '/system', 'setting', 1, 0),
('user', '用户管理', 1, 0, '/user', 'user', 2, 0),
('role', '角色管理', 1, 0, '/role', 'role', 3, 0),
('permission', '权限管理', 1, 0, '/permission', 'lock', 4, 0),
-- 系统管理子菜单
('system:user', '用户管理', 1, 1, '/system/user', 'user', 1, 0),
('system:role', '角色管理', 1, 1, '/system/role', 'role', 2, 0),
('system:permission', '权限管理', 1, 1, '/system/permission', 'lock', 3, 0),
-- 按钮权限
('system:user:add', '用户新增', 2, 5, NULL, NULL, 1, 0),
('system:user:edit', '用户编辑', 2, 5, NULL, NULL, 2, 0),
('system:user:delete', '用户删除', 2, 5, NULL, NULL, 3, 0),
('system:user:view', '用户查看', 2, 5, NULL, NULL, 4, 0),
('system:role:add', '角色新增', 2, 6, NULL, NULL, 1, 0),
('system:role:edit', '角色编辑', 2, 6, NULL, NULL, 2, 0),
('system:role:delete', '角色删除', 2, 6, NULL, NULL, 3, 0),
('system:role:view', '角色查看', 2, 6, NULL, NULL, 4, 0);

-- 1.3 角色-权限关联
-- 超级管理员：所有权限
INSERT INTO `sys_role_permission` (`role_id`, `permission_id`, `created_by`) VALUES
(1, 1, 1), (1, 2, 1), (1, 3, 1), (1, 4, 1),  -- 顶级菜单
(1, 5, 1), (1, 6, 1), (1, 7, 1),              -- 系统管理子菜单
(1, 8, 1), (1, 9, 1), (1, 10, 1), (1, 11, 1), -- 用户管理按钮
(1, 12, 1), (1, 13, 1), (1, 14, 1), (1, 15, 1); -- 角色管理按钮

-- 管理员：部分权限
INSERT INTO `sys_role_permission` (`role_id`, `permission_id`, `created_by`) VALUES
(2, 1, 1), (2, 2, 1), (2, 3, 1),                -- 顶级菜单
(2, 5, 1), (2, 6, 1), (2, 7, 1),                -- 系统管理子菜单
(2, 8, 1), (2, 9, 1), (2, 11, 1),               -- 用户管理：新增、编辑、查看
(2, 12, 1), (2, 13, 1), (2, 15, 1);             -- 角色管理：新增、编辑、查看

-- 普通用户：基本查看权限
INSERT INTO `sys_role_permission` (`role_id`, `permission_id`, `created_by`) VALUES
(3, 2, 1),                                      -- 用户管理顶级菜单
(3, 5, 1),                                      -- 用户管理子菜单
(3, 11, 1);                                     -- 用户查看

-- 1.4 超级管理员账号
INSERT INTO `sys_users` (`username`, `password`, `nickname`, `email`, `phone`, `status`, `is_super_admin`) VALUES
('admin', '8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92', '超级管理员', 'admin@example.com', '13800138000', 0, 1);

-- 1.5 超级管理员关联角色
INSERT INTO `sys_user_role` (`user_id`, `role_id`, `created_by`) VALUES
(1, 1, 1);  -- admin 分配超级管理员角色

-- ==========================================================
-- [2/2] 可选数据 START - 其他测试用户 (user1 / user2 / user3)
-- 用途：多账号测试、权限分级测试
-- 跳过方式：见文件头部说明
-- ==========================================================

-- 2.1 其他测试用户
INSERT INTO `sys_users` (`username`, `password`, `nickname`, `email`, `phone`, `status`, `is_super_admin`) VALUES
('user1', '8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92', '张三', 'zhangsan@example.com', '13800138001', 0, 0),
('user2', '8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92', '李四', 'lisi@example.com', '13800138002', 0, 0),
('user3', '8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92', '王五', 'wangwu@example.com', '13800138003', 0, 0);

-- 2.2 其他测试用户关联角色
INSERT INTO `sys_user_role` (`user_id`, `role_id`, `created_by`) VALUES
(2, 3, 1),  -- user1 分配普通用户角色
(3, 3, 1),  -- user2 分配普通用户角色
(4, 2, 1);  -- user3 分配管理员角色

-- ==========================================================
-- [2/2] 可选数据 END
-- ==========================================================