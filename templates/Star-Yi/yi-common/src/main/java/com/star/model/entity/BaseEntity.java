package com.star.model.entity;

import com.fasterxml.jackson.annotation.JsonFormat;
import jakarta.annotation.Nullable;
import org.babyfish.jimmer.sql.Column;
import org.babyfish.jimmer.sql.LogicalDeleted;
import org.babyfish.jimmer.sql.MappedSuperclass;

import java.time.LocalDateTime;

@MappedSuperclass // (1)
public interface BaseEntity {

    @JsonFormat(pattern = "yyyy-MM-dd HH:mm:ss")
    LocalDateTime createdTime(); // 这是创建时间

    @JsonFormat(pattern = "yyyy-MM-dd HH:mm:ss")
    LocalDateTime modifiedTime(); // 这是修改时间

//     这是删除时间
    @JsonFormat(pattern = "yyyy-MM-dd HH:mm:ss")
    @Nullable
    LocalDateTime deletedTime();

    // 是否删除
    @LogicalDeleted("true")
    @Column(name = "is_deleted")
    boolean deleted();
}
