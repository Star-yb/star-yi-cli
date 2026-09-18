package com.star.yidemo.model.entity;

import com.star.model.entity.BaseEntity;
import jakarta.annotation.Nullable;
import org.babyfish.jimmer.sql.Entity;
import org.babyfish.jimmer.sql.GeneratedValue;
import org.babyfish.jimmer.sql.GenerationType;
import org.babyfish.jimmer.sql.Id;
import org.babyfish.jimmer.sql.Key;
import org.babyfish.jimmer.sql.Table;

/**
 * 测试用文章实体。
 *
 * <pre>
 * CREATE TABLE demo_article (
 *     id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '主键',
 *     created_time DATETIME NOT NULL COMMENT '创建时间',
 *     modified_time DATETIME NOT NULL COMMENT '修改时间',
 *     deleted_time DATETIME NULL COMMENT '删除时间',
 *     is_deleted BIT NOT NULL DEFAULT b'0' COMMENT '逻辑删除',
 *     article_code VARCHAR(64) NOT NULL COMMENT '文章编码',
 *     article_title VARCHAR(128) NOT NULL COMMENT '文章标题',
 *     content VARCHAR(500) NULL COMMENT '文章内容',
 *     UNIQUE KEY uk_demo_article_code (article_code)
 * ) COMMENT='演示文章表';
 * </pre>
 */
@Entity
@Table(name = "demo_article")
public interface DemoArticle extends BaseEntity {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    long id();

    @Key
    String articleCode();

    String articleTitle();

    @Nullable
    String content();
}
