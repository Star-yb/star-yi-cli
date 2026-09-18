package com.star.common.jimmer;


import org.babyfish.jimmer.sql.meta.UserIdGenerator;

import java.net.NetworkInterface;
import java.security.SecureRandom;
import java.time.Instant;
import java.util.Enumeration;

/**
 * Jimmer 专用雪花 ID 生成器
 *
 * 结构: 0(1bit) | 时间戳(41bit) | 工作节点(10bit) | 序列号(12bit) = 64bit Long
 *
 * 特点:
 * - 支持自定义起始时间戳，延长可用年限
 * - 自动获取/配置工作节点ID，避免分布式冲突
 * - 时钟回拨保护（容忍范围内等待，超出则抛异常）
 * - 单节点内支持每毫秒 4096 个 ID 并发
 */
public class SnowflakeIdGenerator implements UserIdGenerator<Long> {

    // ==================== 基础配置 ====================

    /** 起始时间戳 (2024-01-01 00:00:00 UTC)，可用约69年 */
    private static final long START_TIMESTAMP = 1704067200000L;

    /** 工作节点ID位数 */
    private static final long WORKER_ID_BITS = 10L;
    /** 序列号位数 */
    private static final long SEQUENCE_BITS = 12L;

    /** 最大工作节点ID: 1023 */
    private static final long MAX_WORKER_ID = ~(-1L << WORKER_ID_BITS);
    /** 最大序列号: 4095 */
    private static final long MAX_SEQUENCE = ~(-1L << SEQUENCE_BITS);

    /** 工作节点ID左移位数 */
    private static final long WORKER_ID_SHIFT = SEQUENCE_BITS;
    /** 时间戳左移位数 */
    private static final long TIMESTAMP_SHIFT = SEQUENCE_BITS + WORKER_ID_BITS;

    // ==================== 运行状态 ====================

    private final long workerId;
    private volatile long lastTimestamp = -1L;
    private volatile long sequence = 0L;

    // 用于时钟回拨的容忍度（毫秒），超过则抛异常
    private static final long CLOCK_BACKWARD_TOLERANCE_MS = 5;

    /**
     * 自动推断工作节点ID（基于MAC地址或随机数）
     */
    public SnowflakeIdGenerator() {
        this.workerId = generateWorkerId();
    }

    /**
     * 指定工作节点ID（0-1023），适合容器化/K8s环境手动配置
     */
    public SnowflakeIdGenerator(long workerId) {
        if (workerId < 0 || workerId > MAX_WORKER_ID) {
            throw new IllegalArgumentException(
                    String.format("Worker ID must be between 0 and %d, got: %d", MAX_WORKER_ID, workerId)
            );
        }
        this.workerId = workerId;
    }

    /**
     * Jimmer 接口实现：为任意实体生成 ID
     */
    @Override
    public Long generate(Class<?> entityType) {
        return nextId();
    }

    /**
     * 生成下一个雪花 ID（线程安全）
     */
    public synchronized long nextId() {
        long currentTimestamp = getCurrentTimestamp();

        // 处理时钟回拨
        if (currentTimestamp < lastTimestamp) {
            long offset = lastTimestamp - currentTimestamp;
            if (offset <= CLOCK_BACKWARD_TOLERANCE_MS) {
                // 小幅度回拨：等待追赶
                try {
                    Thread.sleep(offset);
                    currentTimestamp = getCurrentTimestamp();
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    throw new RuntimeException("Clock backward wait interrupted", e);
                }
            } else {
                throw new RuntimeException(
                        String.format("Clock moved backwards by %d ms. Refusing to generate ID", offset)
                );
            }
        }

        if (currentTimestamp == lastTimestamp) {
            // 同一毫秒内，序列号递增
            sequence = (sequence + 1) & MAX_SEQUENCE;
            if (sequence == 0) {
                // 序列号溢出，等待下一毫秒
                currentTimestamp = waitUntilNextMillis(lastTimestamp);
            }
        } else {
            // 新的毫秒，序列号重置
            sequence = 0L;
        }

        lastTimestamp = currentTimestamp;

        // 组装64位ID
        return ((currentTimestamp - START_TIMESTAMP) << TIMESTAMP_SHIFT)
                | (workerId << WORKER_ID_SHIFT)
                | sequence;
    }

    // ==================== 辅助方法 ====================

    private long getCurrentTimestamp() {
        return Instant.now().toEpochMilli();
    }

    private long waitUntilNextMillis(long lastTimestamp) {
        long timestamp = getCurrentTimestamp();
        while (timestamp <= lastTimestamp) {
            timestamp = getCurrentTimestamp();
        }
        return timestamp;
    }

    /**
     * 自动生成工作节点ID（基于MAC地址后几位，失败则使用随机数）
     */
    private static long generateWorkerId() {
        try {
            Enumeration<NetworkInterface> interfaces = NetworkInterface.getNetworkInterfaces();
            while (interfaces.hasMoreElements()) {
                NetworkInterface ni = interfaces.nextElement();
                byte[] mac = ni.getHardwareAddress();
                if (mac != null && mac.length >= 6) {
                    // 取MAC后4位，与最大节点ID取模
                    long id = ((mac[mac.length - 2] & 0xFF) << 8)
                            | (mac[mac.length - 1] & 0xFF);
                    return id & MAX_WORKER_ID;
                }
            }
        } catch (Exception e) {
            // 忽略异常，降级为随机数
        }

        // 降级：使用安全随机数
        SecureRandom random = new SecureRandom();
        return random.nextInt((int) MAX_WORKER_ID + 1);
    }

    // ==================== 工具方法（可选） ====================

    /**
     * 解析雪花ID的组成部分（调试用）
     */
    public static IdComponents parse(long id) {
        long timestamp = (id >> TIMESTAMP_SHIFT) + START_TIMESTAMP;
        long worker = (id >> WORKER_ID_SHIFT) & MAX_WORKER_ID;
        long seq = id & MAX_SEQUENCE;
        return new IdComponents(timestamp, worker, seq);
    }

    public record IdComponents(long timestamp, long workerId, long sequence) {
        @Override
        public String toString() {
            return String.format("IdComponents{timestamp=%d, workerId=%d, sequence=%d, date=%s}",
                    timestamp, workerId, sequence, Instant.ofEpochMilli(timestamp));
        }
    }
}