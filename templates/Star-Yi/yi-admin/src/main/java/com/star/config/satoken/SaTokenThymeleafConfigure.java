package com.star.config.satoken;

import cn.dev33.satoken.stp.StpUtil;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.context.annotation.Configuration;
import org.thymeleaf.spring6.view.ThymeleafViewResolver;

@Configuration
public class SaTokenThymeleafConfigure {

    // 为 Thymeleaf 注入全局变量，以便在页面中调用 Sa-Token 的方法
    // <p>调用 StpLogic 方法调用测试</p>
// <p th:if="${stp.isLogin()}">
//     从SaSession中取值：
//     <span th:text="${stp.getSession().get('name')}"></span>
// </p>
    @Autowired
    private void configureThymeleafStaticVars(ThymeleafViewResolver viewResolver) {
        viewResolver.addStaticVariable("stp", StpUtil.stpLogic);
    }
}
