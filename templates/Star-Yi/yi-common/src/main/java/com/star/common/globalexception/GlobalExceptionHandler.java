package com.star.common.globalexception;

import cn.dev33.satoken.exception.NotLoginException;
import cn.dev33.satoken.exception.SaTokenException;
import cn.dev33.satoken.util.SaResult;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import lombok.extern.slf4j.Slf4j;
import org.jetbrains.annotations.NotNull;
import org.springframework.http.converter.HttpMessageNotReadableException;
import org.springframework.web.HttpRequestMethodNotSupportedException;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;
import org.springframework.web.servlet.NoHandlerFoundException;
import org.springframework.web.servlet.resource.NoResourceFoundException;

/**
 * 全局异常处理
 *
 * @author click33
 */
@Slf4j
@RestControllerAdvice
public class GlobalExceptionHandler {

    // 处理 SaTokenException 异常
    @ExceptionHandler(SaTokenException.class)
    public SaResult handlerSaTokenException(@NotNull SaTokenException e, HttpServletRequest request, HttpServletResponse response) {
//		throw new SaTokenException("无效redirect：" ).setCode(1001);


        // 根据不同异常细分状态码返回不同的提示
        if (e.getCode() == 30001) {
            return SaResult.error("redirect 重定向 url 是一个无效地址");
        }

        // 更多 code 码判断 ...
        if (e.getCode() == 11011) {
            return SaResult.notLogin().setMsg("未能读取到有效 token");
        }
        if (e.getCode() == 11012) {
            return SaResult.notLogin().setMsg("提供的token 无效");
        }
        if (e.getCode() == 11013) {
            return SaResult.notLogin().setMsg("提供的token 已过期");
        }
        if (e.getCode() == 11041) {
//            角色不匹配情况
//            超级管理员：无此角色：*
            return SaResult.notLogin().setMsg("角色暂无权限");
        }
        // 默认的提示
        e.printStackTrace();
        return SaResult.error(e.getMessage()).setCode(e.getCode());
    }

    //    NotLoginException 未登录异常
    @ExceptionHandler(NotLoginException.class)
    public SaResult handlerException(NotLoginException e) {
        return SaResult.error(e.getMessage());
    }


    //    HttpRequestMethodNotSupportedException 请求方法不支持
    @ExceptionHandler(HttpRequestMethodNotSupportedException.class)
    public SaResult handlerHttpRequestMethodNotSupportedException(HttpRequestMethodNotSupportedException e, HttpServletRequest request, HttpServletResponse response) {
        return SaResult.error("请求方法不支持").setCode(405);
    }


    //NoHandlerFoundException 未找到处理器
    @ExceptionHandler(NoHandlerFoundException.class)
    public SaResult handlerNoHandlerFoundException(NoHandlerFoundException e, HttpServletRequest request, HttpServletResponse response) {
        return SaResult.error("接口不存在").setCode(404);
    }
//    NoResourceFoundException 未找到资源
    @ExceptionHandler(NoResourceFoundException.class )
    public SaResult handlerNoResourceFoundException(NoResourceFoundException e, HttpServletRequest request, HttpServletResponse response) {
        return SaResult.error("资源不存在").setCode(404);
    }


    //    HttpMessageNotReadableException 缺少请求体
    @ExceptionHandler(HttpMessageNotReadableException.class)
    public SaResult handlerHttpMessageNotReadableException(HttpMessageNotReadableException e, HttpServletRequest request, HttpServletResponse response) {

        log.error(String.valueOf(e.getMostSpecificCause()));
        log.error(String.valueOf(e.getCause()));
        return SaResult.error("缺少请求体或数据不完整").setCode(400);
    }


    // 处理其他普通异常
    @ExceptionHandler(Exception.class)
    public SaResult handlerException(Exception e, HttpServletRequest request, HttpServletResponse response) {
        e.printStackTrace();
        return SaResult.error(e.getMessage());
    }
}


//@RestControllerAdvice
//public class GlobalExceptionHandler {
//	@ExceptionHandler(SaTokenException.class)
//	public SaResult handlerSaTokenException(SaTokenException e, HttpServletRequest request, HttpServletResponse response) {
//		System.out.println(request);
//		System.out.println(response);
//		System.out.println(e.getCode());
//
//		// 根据不同异常细分状态码返回不同的提示
//		if(e.getCode() == 30001) {
//			return SaResult.error("redirect 重定向 url 是一个无效地址");
//		}
//		// 更多 code 码判断 ...
//
//		// 默认的提示
//		e.printStackTrace();
//		return SaResult.error(e.getMessage());
//	}
//}

