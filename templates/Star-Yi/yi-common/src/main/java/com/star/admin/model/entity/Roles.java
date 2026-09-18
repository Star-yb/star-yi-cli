package com.star.admin.model.entity;


import com.star.model.entity.BaseEntity;
import jakarta.annotation.Nullable;
import org.babyfish.jimmer.sql.*;

import java.util.List;

@Entity
@Table(name = "sys_roles")
public interface Roles extends BaseEntity {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    long id();

    @Key
    String roleCode();
    String roleName();

    @Nullable
    String description();

    int status();


    @ManyToMany(mappedBy = "roles")
    List<Users> users();

    @ManyToMany(mappedBy = "roles")
    List<Permissions> permissions();



}
