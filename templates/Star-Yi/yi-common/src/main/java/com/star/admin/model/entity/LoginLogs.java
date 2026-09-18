package com.star.admin.model.entity;

import com.fasterxml.jackson.annotation.JsonFormat;
import jakarta.annotation.Nullable;
import org.babyfish.jimmer.sql.GenerationType;
import org.babyfish.jimmer.sql.Entity;
import org.babyfish.jimmer.sql.ForeignKeyType;
import org.babyfish.jimmer.sql.GeneratedValue;
import org.babyfish.jimmer.sql.Id;
import org.babyfish.jimmer.sql.IdView;
import org.babyfish.jimmer.sql.JoinColumn;
import org.babyfish.jimmer.sql.ManyToOne;
import org.babyfish.jimmer.sql.Table;

import java.time.LocalDateTime;

/**
 * 登录日志实体（只追加、不逻辑删除）。
 *
 * <p>login_type: 1=登录, 2=登出</p>
 * <p>status: 0=成功, 1=失败</p>
 */
@Entity
@Table(name = "sys_login_log")
public interface LoginLogs {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    long id();

    @ManyToOne
    @JoinColumn(name = "user_id", foreignKeyType = ForeignKeyType.FAKE)
    @Nullable
    Users user();

    @IdView("user")
    @Nullable
    Long userId();

    /** 登录时填写的账号（用户名或手机号），失败时也可能无对应用户 */
    String username();

    /** 操作类型 1:登录 2:登出 */
    int loginType();

    /** 状态 0:成功 1:失败 */
    int status();

    @Nullable
    String failReason();

    @Nullable
    String ipAddress();

    @Nullable
    String loginLocation();

    @Nullable
    String browser();

    @Nullable
    String os();

    @Nullable
    String userAgent();

    /** Token 脱敏标识，便于与会话关联排查 */
    @Nullable
    String tokenValue();

    @JsonFormat(pattern = "yyyy-MM-dd HH:mm:ss")
    LocalDateTime loginTime();
}
