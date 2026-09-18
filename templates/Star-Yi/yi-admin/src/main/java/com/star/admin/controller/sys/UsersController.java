package com.star.admin.controller.sys;

import cn.dev33.satoken.annotation.SaCheckRole;
import cn.dev33.satoken.stp.StpUtil;
import cn.dev33.satoken.util.SaResult;
import com.star.admin.model.dto.*;
import com.star.admin.model.entity.Users;
import com.star.admin.model.entity.UsersFetcher;
import com.star.admin.model.entity.UsersTable;
import com.star.admin.service.UsersService;
import com.star.utils.AuthCryptoUtils;
import com.star.common.page.PageQuery;
import com.star.utils.ExcelUtils;
import jakarta.servlet.http.HttpServletResponse;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.io.IOException;
import java.time.format.DateTimeFormatter;
import java.util.List;

/**
 * 用户控制器
 */
@Tag(name = "用户管理", description = "用户 CRUD、个人资料、角色绑定与导出，路径 /users")
@RestController
@RequestMapping("/users")
@RequiredArgsConstructor
public class UsersController {

    private final UsersService usersService;

    @Operation(summary = "分页查询用户列表")
    @GetMapping
    public SaResult getUserList(
            PageQuery pageQuery,
            UserSpecification usersSpecification
    ) {
        return SaResult.data(usersService.listUsers(pageQuery, usersSpecification));
    }

    @Operation(summary = "创建用户", description = "默认密码 123456（已哈希）")
    @PostMapping
    public SaResult createUser(@RequestBody UserCreateInputView user) {
        user.setPassword(AuthCryptoUtils.hash("123456"));// Set a default password
        return SaResult.data(usersService.createUser(user));
    }

    @Operation(summary = "更新用户")
    @PutMapping("/{id}")
    public SaResult updateUser(@PathVariable Long id, @RequestBody UserUpdateInputView user) {
        user.setId(id);
        return SaResult.data(usersService.updateUser(user));
    }

    @Operation(summary = "删除用户", description = "逻辑删除")
    @DeleteMapping("/{id}")
    public SaResult deleteUser(@PathVariable Long id) {
        usersService.deleteUser(id);
        return SaResult.ok("删除成功");
    }

    @Operation(summary = "用户详情")
    @GetMapping("/{id}")
    public SaResult getUserById(@PathVariable Long id) {
        return SaResult.data(usersService.obtainUser(id));
    }


    @Operation(summary = "当前登录用户资料")
    @GetMapping("/myInfo")
    public SaResult myInfo() {
        StpUtil.checkLogin();  // 如果当前未登录，这句代码会直接抛出异常 `NotLoginException`
        Users userById = usersService.obtainUser(StpUtil.getLoginIdAsLong());
        return SaResult.data(userById);
    }

    @Operation(summary = "修改当前用户密码")
    @PostMapping("/changePassword")
    public SaResult changePassword(@RequestBody UserPasswordUpdateInputView userPasswordUpdateInputView) {
        usersService.changePassword(StpUtil.getLoginIdAsLong(), userPasswordUpdateInputView.getOldPassword(), userPasswordUpdateInputView.getNewPassword());
        return SaResult.ok("修改成功");
    }

    @Operation(summary = "启用用户")
    @PostMapping("/{id}/enable")
    public SaResult enableStatus(@PathVariable Long id) {
        usersService.enableUser(id);
        return SaResult.ok("修改成功");
    }

    @Operation(summary = "禁用用户")
    @PostMapping("/{id}/disable")
    public SaResult disableStatus(@PathVariable Long id) {
        usersService.disableUser(id);
        return SaResult.ok("修改成功");
    }

    @Operation(summary = "查询用户已绑定角色")
    @GetMapping("/roles/{id}")
    public SaResult getRoleUsers(@PathVariable Long id) {
        return SaResult.data(usersService.getRoleUsers(id));
    }

    @Operation(summary = "更新用户角色分配")
    @PutMapping("/{id}/roles")
    public SaResult updateRole(@PathVariable Long id, @RequestBody UserRoleUpdateInputView userRoleUpdateInputView) {
        return SaResult.data(usersService.updateRole(userRoleUpdateInputView));
    }

    @Operation(summary = "重置用户密码", description = "需超级管理员角色")
    @SaCheckRole("*")
    @PostMapping("/{id}/resetPassword")
    public SaResult resetPassword(@PathVariable Long id) {
        usersService.resetPassword(id);
        return SaResult.ok("修改成功");
    }

    @Operation(summary = "导出用户 Excel", description = "需超级管理员角色；直接下载文件流")
    @SaCheckRole("*")
    @GetMapping("/export")
    public void exportUsers(HttpServletResponse response) throws IOException {
        List<Users> users = usersService.listUsers(null);

        DateTimeFormatter fmt = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");
        List<ExcelUtils.Column<Users>> columns = List.of(
                new ExcelUtils.Column<>("用户ID", Users::id),
                new ExcelUtils.Column<>("用户名", Users::username),
                new ExcelUtils.Column<>("昵称", u -> u.nickname() == null ? "" : u.nickname()),
                new ExcelUtils.Column<>("邮箱", u -> u.email() == null ? "" : u.email()),
                new ExcelUtils.Column<>("手机号", u -> u.phone() == null ? "" : u.phone()),
                new ExcelUtils.Column<>("状态", u -> toStatusText(u.status())),
                new ExcelUtils.Column<>("是否超管", u -> u.superAdmin() == 1 ? "是" : "否"),
                new ExcelUtils.Column<>("最后登录时间", u -> u.lastLoginTime() == null ? "" : u.lastLoginTime().format(fmt))
        );
        ExcelUtils.writeByColumnsResponse(response, "users-export", "用户列表", columns, users);

      /*  ExcelUtils.writeByColumnsPagedResponse(
                response,
                "users-export-paged",
                "用户列表",
                columns,
                1000,
                (pageNo, pageSize) -> {
                    Page<Users> page = usersDao.queryPage(
                            PageRequest.of(pageNo - 1, pageSize),
                            (org.babyfish.jimmer.Specification<Users>) null,
                            t -> t.fetch(UsersFetcher.$.allScalarFields())
                    );
                    return page.getContent();
                }
        );*/
    }

    private String toStatusText(int status) {
        return switch (status) {
            case 0 -> "正常";
            case 1 -> "禁用";
            case 2 -> "删除";
            default -> String.valueOf(status);
        };
    }
}
