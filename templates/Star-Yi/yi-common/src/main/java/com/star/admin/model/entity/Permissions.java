package com.star.admin.model.entity;


import com.star.model.entity.BaseEntity;
import jakarta.annotation.Nullable;
import org.babyfish.jimmer.sql.*;

import java.util.List;

@Entity
@Table(name = "sys_permissions")
public interface Permissions extends BaseEntity {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    long id();
    @Key
    String permissionCode();
    String permissionName();

    int permissionType();

    @ManyToOne()
    @JoinColumn(name = "parent_id",foreignKeyType = ForeignKeyType.FAKE)
    @Nullable
    Permissions parent();

    @OneToMany(mappedBy = "parent")
    List<Permissions> childPermissions();

    @Nullable
    String path();
    @Nullable
    String icon();
    int sortOrder();

    int status();

    @ManyToMany
    @JoinTable(
            name = "sys_role_permission",
            joinColumnName = "permission_id",
            inverseJoinColumnName = "role_id"
    )
    List<Roles> roles();


    @IdView(value = "parent")
    Long parentId();


}
