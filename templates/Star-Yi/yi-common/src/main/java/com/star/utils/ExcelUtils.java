package com.star.utils;

import com.alibaba.excel.EasyExcel;
import com.alibaba.excel.ExcelWriter;
import com.alibaba.excel.write.metadata.WriteSheet;
import jakarta.servlet.http.HttpServletResponse;

import java.io.IOException;
import java.io.OutputStream;
import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.ArrayList;
import java.util.List;
import java.util.Objects;
import java.util.function.Function;

/**
 * EasyExcel 通用导出工具。
 * 支持：
 * 1) 动态列头 + 二维数据
 * 2) 列定义 + 提取器（避免业务层手写二维数组）
 * 3) 分页流式写入（大数据量）
 */
public final class ExcelUtils {

    private ExcelUtils() {
    }

    /**
     * 列定义：列名 + 从对象提取单元格值的函数。
     */
    public record Column<T>(String header, Function<T, Object> extractor) {
        public Column {
            header = header == null ? "" : header;
            extractor = Objects.requireNonNull(extractor, "extractor");
        }
    }

    /**
     * 分页数据提供器：按页返回数据，返回空列表表示结束。
     */
    @FunctionalInterface
    public interface PageFetcher<T> {
        List<T> fetch(int pageNo, int pageSize);
    }

    /**
     * 将一维列名转换为 EasyExcel 需要的二维 head 结构。
     */
    public static List<List<String>> toHead(List<String> columns) {
        Objects.requireNonNull(columns, "columns");
        List<List<String>> head = new ArrayList<>(columns.size());
        for (String column : columns) {
            List<String> one = new ArrayList<>(1);
            one.add(column == null ? "" : column);
            head.add(one);
        }
        return head;
    }

    /**
     * 导出到输出流。
     */
    public static void writeDynamic(OutputStream outputStream, String sheetName, List<String> columns, List<List<Object>> rows) {
        Objects.requireNonNull(outputStream, "outputStream");
        List<List<String>> head = toHead(columns);
        String finalSheetName = (sheetName == null || sheetName.isBlank()) ? "Sheet1" : sheetName;
        List<List<Object>> safeRows = normalizeRows(rows, head.size());
        EasyExcel.write(outputStream).head(head).sheet(finalSheetName).doWrite(safeRows);
    }

    /**
     * 导出到本地文件（xlsx）。
     */
    public static void writeDynamic(Path path, String sheetName, List<String> columns, List<List<Object>> rows) throws IOException {
        Objects.requireNonNull(path, "path");
        try (OutputStream outputStream = Files.newOutputStream(path)) {
            writeDynamic(outputStream, sheetName, columns, rows);
        }
    }

    /**
     * 导出到浏览器响应（xlsx 下载）。
     */
    public static void writeDynamicResponse(
            HttpServletResponse response,
            String rawFileName,
            String sheetName,
            List<String> columns,
            List<List<Object>> rows
    ) throws IOException {
        Objects.requireNonNull(response, "response");
        String finalFileName = (rawFileName == null || rawFileName.isBlank()) ? "export" : rawFileName;
        String encoded = URLEncoder.encode(finalFileName + ".xlsx", StandardCharsets.UTF_8).replace("+", "%20");

        response.setContentType("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet");
        response.setCharacterEncoding(StandardCharsets.UTF_8.name());
        response.setHeader("Content-Disposition", "attachment; filename=\"" + encoded + "\"; filename*=UTF-8''" + encoded);

        writeDynamic(response.getOutputStream(), sheetName, columns, rows);
        response.flushBuffer();
    }

    /**
     * 按列定义导出（适用于普通列表，不需要业务层手工循环拼二维数组）。
     */
    public static <T> void writeByColumns(
            OutputStream outputStream,
            String sheetName,
            List<Column<T>> columns,
            List<T> data
    ) {
        Objects.requireNonNull(columns, "columns");
        List<String> headers = columns.stream().map(Column::header).toList();
        List<List<Object>> rows = mapRows(columns, data);
        writeDynamic(outputStream, sheetName, headers, rows);
    }

    /**
     * 按列定义导出到浏览器下载。
     */
    public static <T> void writeByColumnsResponse(
            HttpServletResponse response,
            String rawFileName,
            String sheetName,
            List<Column<T>> columns,
            List<T> data
    ) throws IOException {
        prepareDownloadResponse(response, rawFileName);
        writeByColumns(response.getOutputStream(), sheetName, columns, data);
        response.flushBuffer();
    }

    /**
     * 分页流式导出（避免一次性加载全部数据到内存）。
     */
    public static <T> void writeByColumnsPagedResponse(
            HttpServletResponse response,
            String rawFileName,
            String sheetName,
            List<Column<T>> columns,
            int pageSize,
            PageFetcher<T> pageFetcher
    ) throws IOException {
        Objects.requireNonNull(response, "response");
        Objects.requireNonNull(columns, "columns");
        Objects.requireNonNull(pageFetcher, "pageFetcher");
        if (pageSize <= 0) {
            throw new IllegalArgumentException("pageSize must be > 0");
        }

        prepareDownloadResponse(response, rawFileName);
        String finalSheetName = (sheetName == null || sheetName.isBlank()) ? "Sheet1" : sheetName;
        List<List<String>> head = toHead(columns.stream().map(Column::header).toList());

        try (ExcelWriter writer = EasyExcel.write(response.getOutputStream()).build()) {
            WriteSheet writeSheet = EasyExcel.writerSheet(finalSheetName).head(head).build();
            for (int pageNo = 1; ; pageNo++) {
                List<T> pageData = pageFetcher.fetch(pageNo, pageSize);
                if (pageData == null || pageData.isEmpty()) {
                    break;
                }
                writer.write(mapRows(columns, pageData), writeSheet);
            }
        }
        response.flushBuffer();
    }

    /**
     * 归一化每行列数：短行补 null，长行截断，避免列错位。
     */
    private static List<List<Object>> normalizeRows(List<List<Object>> rows, int expectedSize) {
        List<List<Object>> result = new ArrayList<>();
        if (rows == null) {
            return result;
        }
        for (List<Object> row : rows) {
            List<Object> safeRow = new ArrayList<>(expectedSize);
            if (row != null) {
                int copy = Math.min(row.size(), expectedSize);
                for (int i = 0; i < copy; i++) {
                    safeRow.add(row.get(i));
                }
            }
            while (safeRow.size() < expectedSize) {
                safeRow.add(null);
            }
            result.add(safeRow);
        }
        return result;
    }

    private static void prepareDownloadResponse(HttpServletResponse response, String rawFileName) {
        Objects.requireNonNull(response, "response");
        String finalFileName = (rawFileName == null || rawFileName.isBlank()) ? "export" : rawFileName;
        String encoded = URLEncoder.encode(finalFileName + ".xlsx", StandardCharsets.UTF_8).replace("+", "%20");
        response.setContentType("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet");
        response.setCharacterEncoding(StandardCharsets.UTF_8.name());
        response.setHeader("Content-Disposition", "attachment; filename=\"" + encoded + "\"; filename*=UTF-8''" + encoded);
    }

    private static <T> List<List<Object>> mapRows(List<Column<T>> columns, List<T> data) {
        List<List<Object>> rows = new ArrayList<>();
        if (data == null || data.isEmpty()) {
            return rows;
        }
        for (T item : data) {
            List<Object> row = new ArrayList<>(columns.size());
            for (Column<T> column : columns) {
                row.add(column.extractor().apply(item));
            }
            rows.add(row);
        }
        return rows;
    }
}
