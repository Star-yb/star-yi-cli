package com.star.admin.controller.webpage;


import cn.dev33.satoken.annotation.SaIgnore;
import cn.dev33.satoken.stp.StpUtil;
import cn.dev33.satoken.util.SaResult;
import com.star.admin.model.dto.*;
import com.star.admin.model.entity.LoginLogs;
import com.star.admin.model.entity.Permissions;
import com.star.admin.model.entity.Roles;
import com.star.admin.model.entity.Users;
import com.star.admin.service.AuthService;
import com.star.admin.support.LoginLogSupport;
import com.star.admin.support.SaAuthLoginSupport;
import com.star.admin.service.PermissionsService;
import com.star.admin.service.RolesService;
import com.star.admin.service.UsersService;
import com.star.common.exception.BusinessException;
import com.star.common.page.PageQuery;
import com.star.common.page.PageResult;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Controller;
import org.springframework.ui.Model;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.servlet.mvc.support.RedirectAttributes;

import java.util.Arrays;
import java.util.List;
import java.util.stream.Collectors;


@Tag(name = "超级管理员-后台页面", description = "Thymeleaf 管理后台页面与维护接口，路径 /admin")
@Slf4j
@Controller
@RequestMapping("/admin")
@RequiredArgsConstructor
public class WebController {

    private final AuthService authService;
    private final UsersService usersService;
    private final RolesService rolesService;
    private final PermissionsService permissionsService;
    private final LoginLogSupport loginLogSupport;

    @Value("${springdoc.swagger-ui.path:/swagger-ui.html}")
    private String swaggerUiPath;

//   转到登录页面

    @Operation(summary = "登录页", description = "返回 Thymeleaf 登录页面，免鉴权")
    @SaIgnore
    @GetMapping("/login")
    public String login() {
        return "login";
    }

    @Operation(summary = "表单登录", description = "Thymeleaf 表单提交登录，成功后重定向到用户管理页")
    @SaIgnore
    @PostMapping("/login")
    public String login(@RequestParam("username") String username,
                        @RequestParam("password") String password,
                        RedirectAttributes redirectAttributes) {
        try {
            Users login = authService.login(username, password);
            SaAuthLoginSupport.establishAdminWebSession(login.id());
            authService.onLoginSuccess(login.id());
            StpUtil.getSession().set("name", login.username());

            return "redirect:/admin/user";
        } catch (BusinessException e) {
            log.warn("后台登录失败: username={}, reason={}", username, e.getMessage());
            redirectAttributes.addFlashAttribute("errorMsg", e.getMessage());
            redirectAttributes.addFlashAttribute("historyUsername", username);
            return "redirect:/admin/login";
        }
    }

    @Operation(summary = "用户管理页", description = "返回用户列表 Thymeleaf 页面")
    @GetMapping("/user")
    public String user(PageQuery pageQuery,
                       UserSpecification usersSpecification,
                       Model model) {

        if (pageQuery.getPage() < 1) {
            pageQuery.setPage(1);
        }
        PageResult<Users> usersPageResult = usersService.listUsers(pageQuery, usersSpecification);
        List<Roles> rolesViews = rolesService.listRoles(null);

        model.addAttribute("usersPageResult", usersPageResult);
        model.addAttribute("rolesList", rolesViews);
        return "sys/user";
    }


    @Operation(summary = "退出登录", description = "仅注销当前 Thymeleaf 超管 Cookie 会话，不影响同账号其他端与 API 请求头 token")
    @SaIgnore
    @GetMapping("/logout")
    public String logout() {
        if (StpUtil.isLogin()) {
            String tokenValue = StpUtil.getTokenValue();
            // 按 token 注销，避免 logout(loginId, device) 踢掉同账号其它浏览器的超管会话
            if (tokenValue != null && !tokenValue.isBlank()) {
                StpUtil.logoutByTokenValue(tokenValue);
            }
        }
        return "redirect:/admin/login?logout=true";
    }
    @Operation(summary = "角色管理页")
    @GetMapping("/role")
    public String role(PageQuery pageQuery,
                       RoleSpecification roleSpecification,
                       Model model) {

        if (pageQuery.getPage() < 1) {
            pageQuery.setPage(1);
        }
        PageResult<Roles> rolesPageResult = rolesService.listRoles(pageQuery, roleSpecification);

        model.addAttribute("rolesPageResult", rolesPageResult);
        return "sys/role";
    }

    @Operation(summary = "角色权限配置页", description = "按 roleId 展示角色与权限勾选界面")
    @GetMapping("/role-permission")
    public String rolePermission(@RequestParam("roleId") Long roleId,
                                 Model model) {
        Roles role = rolesService.obtainRole(roleId);
        List<Permissions> permissionRoles = permissionsService.getPermissionRoles(roleId);
        List<Permissions> permissionsList = permissionsService.listPermissions(null);

        model.addAttribute("role", role);
        model.addAttribute("permissionsList", permissionsList);
        model.addAttribute("permissionRoles", permissionRoles);
        return "sys/role-permission";
    }

    @Operation(summary = "权限树管理页")
    @GetMapping("/permission")
    public String permission(PageQuery pageQuery,
                             PermissionSpecification permissionSpecification,
                             Model model) {

//        if (pageQuery.getPage() < 1) {
//            pageQuery.setPage(1);
//        }

        PageResult<Permissions> permissionsPageResult = permissionsService.getPermissionsTree(pageQuery, permissionSpecification);

        model.addAttribute("permissionsPageResult", permissionsPageResult);
        return "sys/permission";
    }

    @Operation(summary = "权限列表页", description = "平铺列表视图")
    @GetMapping("/permission-list")
    public String permissionList(PageQuery pageQuery, PermissionSpecification permissionSpecification,
                                 Model model) {
        if (pageQuery.getPage() < 1) {
            pageQuery.setPage(1);
        }
        PageResult<Permissions> permissionsPageResult = permissionsService.listPermissions(pageQuery, permissionSpecification);
        model.addAttribute("permissionsPageResult", permissionsPageResult);
        return "sys/permission-list";
    }

    @Operation(summary = "接口文档页", description = "iframe 内嵌 Swagger UI，工具栏可新标签页打开")
    @GetMapping("/api-docs")
    public String apiDocs(Model model) {
        model.addAttribute("swaggerUiPath", swaggerUiPath);
        return "sys/api-docs";
    }

    @Operation(summary = "登录日志页", description = "支持按用户名、登录类型、状态筛选")
    @GetMapping("/login-log")
    public String loginLog(
            PageQuery pageQuery,
            @RequestParam(required = false) String username,
            @RequestParam(required = false) Integer loginType,
            @RequestParam(required = false) Integer status,
            Model model
    ) {
        if (pageQuery.getPage() < 1) {
            pageQuery.setPage(1);
        }
        PageResult<LoginLogs> loginLogsPageResult = loginLogSupport.list(pageQuery, username, loginType, status);
        model.addAttribute("loginLogsPageResult", loginLogsPageResult);
        model.addAttribute("totalLogCount", loginLogSupport.count());
        return "sys/login-log";
    }

    @Operation(summary = "删除单条登录日志")
    @DeleteMapping("/login-log/{id}")
    @ResponseBody
    public SaResult deleteLoginLog(@PathVariable Long id) {
        int rows = loginLogSupport.deleteByIds(List.of(id));
        return SaResult.ok("已删除 " + rows + " 条记录");
    }

    @Operation(summary = "批量删除登录日志", description = "ids 为逗号分隔的 ID 列表")
    @DeleteMapping("/login-log/batch")
    @ResponseBody
    public SaResult batchDeleteLoginLog(@RequestParam("ids") String ids) {
        List<Long> idList = Arrays.stream(ids.split(","))
                .map(String::trim)
                .filter(s -> !s.isEmpty())
                .map(Long::parseLong)
                .collect(Collectors.toList());
        int rows = loginLogSupport.deleteByIds(idList);
        return SaResult.ok("已删除 " + rows + " 条记录");
    }

    @Operation(summary = "清理 N 天前的登录日志")
    @DeleteMapping("/login-log/clean/older-than")
    @ResponseBody
    public SaResult cleanLoginLogOlderThan(@RequestParam int days) {
        int rows = loginLogSupport.deleteOlderThanDays(days);
        return SaResult.ok("已清理 " + rows + " 条（" + days + " 天前）");
    }

    @Operation(summary = "保留最新 N 条登录日志", description = "删除其余记录")
    @DeleteMapping("/login-log/clean/keep-latest")
    @ResponseBody
    public SaResult cleanLoginLogKeepLatest(@RequestParam int count) {
        int rows = loginLogSupport.deleteKeepLatest(count);
        return SaResult.ok("已清理 " + rows + " 条，保留最新 " + count + " 条");
    }

    @Operation(summary = "清空全部登录日志")
    @DeleteMapping("/login-log/clean/all")
    @ResponseBody
    public SaResult cleanLoginLogAll() {
        int rows = loginLogSupport.deleteAll();
        return SaResult.ok("已清空全部 " + rows + " 条记录");
    }


    @Operation(summary = "物理删除用户", description = "后台管理用，不可恢复")
    @DeleteMapping("/user/{id}")
    @ResponseBody
    public SaResult deleteUser(@PathVariable("id") Long id) {
        usersService.adminDelete(id);
        return SaResult.ok("删除成功");

    }

    @Operation(summary = "物理删除角色", description = "后台管理用，不可恢复")
    @DeleteMapping("/role/{id}")
    @ResponseBody
    public SaResult deleteRole(@PathVariable("id") Long id) {
        rolesService.adminDelete(id);
        return SaResult.ok("删除成功");

    }

    @Operation(summary = "物理删除权限", description = "后台管理用，不可恢复")
    @DeleteMapping("/permission/{id}")
    @ResponseBody
    public SaResult deletePermission(@PathVariable("id") Long id) {
        permissionsService.adminDelete(id);
        return SaResult.ok("删除成功");
    }
}
