package com.star.common.jimmer;

import org.babyfish.jimmer.sql.meta.UserIdGenerator;

import java.net.NetworkInterface;
import java.security.SecureRandom;
import java.time.Instant;
import java.util.Enumeration;

/**
 * Jimmer 专用：可校验有序 Long ID 生成器（VOID）。
 *
 * <p>替代 {@link SnowflakeIdGenerator}，在保持 {@code long} 类型的前提下：</p>
 * <ul>
 *   <li>ID 严格落在 JavaScript {@code Number.MAX_SAFE_INTEGER} 以内，前后端 JSON 无损</li>
 *   <li>按秒级时间戳单调递增，利于 B+ 树索引与业务排序</li>
 *   <li>内嵌 4 位校验码，可配合 {@link VerifiableOrderedId#isValid(long)} 做拦截校验</li>
 * </ul>
 *
 * <p>位布局详见 {@link VerifiableOrderedId}。</p>
 */
public class VerifiableOrderedIdGenerator implements UserIdGenerator<Long> {

    private static final int MAX_WORKER_ID = (1 << 7) - 1;
    private static final int MAX_SEQUENCE = (1 << 9) - 1;
    private static final long CLOCK_BACKWARD_TOLERANCE_MS = 5L;

    private final int workerId;
    private volatile long lastEpochSecond = -1L;
    private volatile int sequence = 0;

    public VerifiableOrderedIdGenerator() {
        this.workerId = generateWorkerId();
    }

    /**
     * 显式指定工作节点（0–127），适合容器/K8s 环境。
     */
    public VerifiableOrderedIdGenerator(int workerId) {
        if (workerId < 0 || workerId > MAX_WORKER_ID) {
            throw new IllegalArgumentException(
                    "Worker ID must be between 0 and " + MAX_WORKER_ID + ", got: " + workerId
            );
        }
        this.workerId = workerId;
    }

    @Override
    public Long generate(Class<?> entityType) {
        return nextId();
    }

    /**
     * 生成下一个 VOID ID（线程安全）。
     */
    public synchronized long nextId() {
        long currentSecond = Instant.now().getEpochSecond();

        if (currentSecond < lastEpochSecond) {
            long offsetMs = (lastEpochSecond - currentSecond) * 1000L;
            if (offsetMs <= CLOCK_BACKWARD_TOLERANCE_MS) {
                try {
                    Thread.sleep(offsetMs);
                    currentSecond = Instant.now().getEpochSecond();
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    throw new RuntimeException("Clock backward wait interrupted", e);
                }
            } else {
                throw new RuntimeException(
                        "Clock moved backwards by " + offsetMs + " ms, refusing to generate VOID"
                );
            }
        }

        if (currentSecond == lastEpochSecond) {
            sequence = (sequence + 1) & MAX_SEQUENCE;
            if (sequence == 0) {
                currentSecond = waitUntilNextSecond(lastEpochSecond);
            }
        } else {
            sequence = 0;
        }

        lastEpochSecond = currentSecond;
        return VerifiableOrderedId.compose(
                VerifiableOrderedId.CURRENT_VERSION,
                currentSecond,
                workerId,
                sequence
        );
    }

    private long waitUntilNextSecond(long lastSecond) {
        long second = Instant.now().getEpochSecond();
        while (second <= lastSecond) {
            second = Instant.now().getEpochSecond();
        }
        return second;
    }

    private static int generateWorkerId() {
        try {
            Enumeration<NetworkInterface> interfaces = NetworkInterface.getNetworkInterfaces();
            while (interfaces.hasMoreElements()) {
                NetworkInterface ni = interfaces.nextElement();
                byte[] mac = ni.getHardwareAddress();
                if (mac != null && mac.length >= 6) {
                    int id = ((mac[mac.length - 2] & 0xFF) << 8) | (mac[mac.length - 1] & 0xFF);
                    return id & MAX_WORKER_ID;
                }
            }
        } catch (Exception ignored) {
            // 降级为随机节点
        }
        return new SecureRandom().nextInt(MAX_WORKER_ID + 1);
    }
}
