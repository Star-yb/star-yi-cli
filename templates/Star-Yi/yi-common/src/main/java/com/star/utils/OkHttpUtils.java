package com.star.utils;

import okhttp3.FormBody;
import okhttp3.HttpUrl;
import okhttp3.MediaType;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.RequestBody;
import okhttp3.Response;
import okhttp3.ResponseBody;

import java.io.IOException;
import java.io.UncheckedIOException;
import java.nio.charset.StandardCharsets;
import java.util.Map;
import java.util.Objects;
import java.util.concurrent.TimeUnit;

/**
 * OkHttp 同步请求轻量封装，适用于简单 GET / JSON / 表单场景。
 */
public final class OkHttpUtils {

    private static final MediaType JSON_MEDIA_TYPE = MediaType.parse("application/json; charset=utf-8");

    private static final OkHttpClient DEFAULT_CLIENT = new OkHttpClient.Builder()
            .connectTimeout(30, TimeUnit.SECONDS)
            .readTimeout(30, TimeUnit.SECONDS)
            .writeTimeout(30, TimeUnit.SECONDS)
            .callTimeout(60, TimeUnit.SECONDS)
            .build();

    private OkHttpUtils() {
    }

    /**
     * 默认客户端，可按需在外层用 {@link OkHttpClient#newBuilder()} 派生。
     */
    public static OkHttpClient client() {
        return DEFAULT_CLIENT;
    }

    public static HttpResult get(String url) {
        return get(url, Map.of());
    }

    public static HttpResult get(String url, Map<String, String> headers) {
        Request.Builder builder = new Request.Builder().url(Objects.requireNonNull(url, "url")).get();
        headers.forEach(builder::addHeader);
        return execute(DEFAULT_CLIENT, builder.build());
    }

    /**
     * GET，查询参数附加在 url 上（会对 value 编码）。
     */
    public static HttpResult get(String url, Map<String, String> queryParams, Map<String, String> headers) {
        HttpUrl base = Objects.requireNonNull(HttpUrl.parse(url), "invalid url: " + url);
        HttpUrl.Builder ub = base.newBuilder();
        if (queryParams != null) {
            queryParams.forEach((k, v) -> {
                if (k != null && v != null) {
                    ub.addQueryParameter(k, v);
                }
            });
        }
        Request.Builder builder = new Request.Builder().url(ub.build()).get();
        if (headers != null) {
            headers.forEach(builder::addHeader);
        }
        return execute(DEFAULT_CLIENT, builder.build());
    }

    public static HttpResult postJson(String url, String json) {
        return postJson(url, json, Map.of());
    }

    public static HttpResult postJson(String url, String json, Map<String, String> headers) {
        RequestBody body = RequestBody.create(json == null ? "" : json, JSON_MEDIA_TYPE);
        Request.Builder builder = new Request.Builder()
                .url(Objects.requireNonNull(url, "url"))
                .post(body);
        if (headers != null) {
            headers.forEach(builder::addHeader);
        }
        return execute(DEFAULT_CLIENT, builder.build());
    }

    public static HttpResult postForm(String url, Map<String, String> formFields) {
        return postForm(url, formFields, Map.of());
    }

    public static HttpResult postForm(String url, Map<String, String> formFields, Map<String, String> headers) {
        FormBody.Builder form = new FormBody.Builder(StandardCharsets.UTF_8);
        if (formFields != null) {
            formFields.forEach(form::add);
        }
        Request.Builder builder = new Request.Builder()
                .url(Objects.requireNonNull(url, "url"))
                .post(form.build());
        if (headers != null) {
            headers.forEach(builder::addHeader);
        }
        return execute(DEFAULT_CLIENT, builder.build());
    }

    public static HttpResult execute(Request request) {
        return execute(DEFAULT_CLIENT, request);
    }

    public static HttpResult execute(OkHttpClient client, Request request) {
        try (Response response = client.newCall(Objects.requireNonNull(request, "request")).execute()) {
            int code = response.code();
            ResponseBody responseBody = response.body();
            String text = responseBody == null ? "" : responseBody.string();
            return new HttpResult(code, text);
        } catch (IOException e) {
            throw new UncheckedIOException(e);
        }
    }
}
