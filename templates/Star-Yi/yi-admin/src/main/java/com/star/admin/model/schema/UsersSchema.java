package com.star.admin.model.schema;

import com.star.admin.model.entity.Users;
import com.star.admin.model.entity.UsersFetcher;
import org.babyfish.jimmer.sql.fetcher.Fetcher;

/**
 * 统一声明命名的 Fetcher 常量，供 @FetchBy 使用以及服务端查询选择字段
 */
public interface UsersSchema {

    // 登录请求接收的字段（输入 DTO 形状）
    Fetcher<Users> LOGIN_REQUEST = UsersFetcher.$
            .username()
            .password();

    // 登录响应返回的字段（输出 DTO 形状）
    Fetcher<Users> LOGIN_RESPONSE = UsersFetcher.$
            .username();

//    修改密码
    Fetcher<Users> UPDATE_PASSWORD = UsersFetcher.$
            .password();


    //    添加/修改通用字段
    Fetcher<Users> COMMON = UsersFetcher.$
            .allScalarFields()
            .password(false);

}