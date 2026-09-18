package com.star.admin.model.entity;

import com.fasterxml.jackson.annotation.JsonFormat;
import com.star.common.jimmer.VerifiableOrderedIdGenerator;
import com.star.model.entity.BaseEntity;
import jakarta.annotation.Nullable;
import org.babyfish.jimmer.sql.*;

import java.time.LocalDateTime;
import java.util.List;

@Entity
@Table(name = "sys_users")
public interface Users extends BaseEntity {

//    @Id
//    @GeneratedValue(strategy = GenerationType.IDENTITY)
//    long id();
    @Id
    @GeneratedValue(generatorType = VerifiableOrderedIdGenerator.class)
    long id();


//    用户名 密码 真实姓名 邮箱 电话
    @Key
    String username();


    String password();

    @Nullable
    String nickname();

    @Nullable
    String email();

    @Nullable
    String phone();

    //    状态 0:正常 1:禁用 2:删除
    int status();


//    是否超级管理员 0:否 1:是
    @Column(name = "is_super_admin")
    int superAdmin();

//    最后登录时间
    @JsonFormat(pattern = "yyyy-MM-dd HH:mm:ss")
    @Nullable
    LocalDateTime lastLoginTime();


    @ManyToMany
    @JoinTable(
            name = "sys_user_role",
            joinColumnName = "user_id",
            inverseJoinColumnName = "role_id"
    )
    List<Roles> roles();

    @IdView(value = "roles")
    List<Long> roleIds();

}