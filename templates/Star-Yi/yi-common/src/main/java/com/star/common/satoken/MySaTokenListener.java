package com.star.common.satoken;

import cn.dev33.satoken.listener.SaTokenListenerForSimple;
import cn.dev33.satoken.stp.parameter.SaLoginParameter;
import org.springframework.stereotype.Component;

/**
 * 自定义侦听器的实现
 */
@Component
public class MySaTokenListener extends SaTokenListenerForSimple {

    /** 每次登录时触发 */
    @Override
    public void doLogin(String loginType, Object loginId, String tokenValue, SaLoginParameter loginParameter) {
        System.out.println("---------- 自定义侦听器实现 doLogin");
    }


}
