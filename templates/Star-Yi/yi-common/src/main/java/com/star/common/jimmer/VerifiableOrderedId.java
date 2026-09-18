package com.star.common.jimmer;

import java.time.Instant;
import java.util.Optional;

/**
 * 可校验有序 Long ID（VOID）的编解码与合法性校验。
 *
 * <p>位布局（共 52 位有效载荷，最大值 4503599627370495，严格小于
 * {@link #JS_MAX_SAFE_INTEGER}，JSON 经 JavaScript 解析无精度损失）：</p>
 * <pre>
 * | VER(3) | TS_SEC(29) | WORKER(7) | SEQ(9) | CHK(4) |
 * </pre>
 *
 * <ul>
 *   <li>VER：格式版本，当前固定为 1</li>
 *   <li>TS_SEC：自 {@link #EPOCH_SECOND} 起的秒级时间戳（单调递增，约 17 年窗口）</li>
 *   <li>WORKER：工作节点（0–127）</li>
 *   <li>SEQ：同一秒内序列号（0–511）</li>
 *   <li>CHK：校验位，由版本/时间/节点/序列与盐值混合得出，供网关或拦截器快速拒收伪造 ID</li>
 * </ul>
 */
public final class VerifiableOrderedId {

    /** JavaScript Number.MAX_SAFE_INTEGER，生成 ID 必须低于此值 */
    public static final long JS_MAX_SAFE_INTEGER = 9007199254740991L;

    /** 当前 ID 格式版本 */
    public static final int CURRENT_VERSION = 1;

    /** 纪元：2024-01-01 00:00:00 UTC（与旧雪花生成器起点一致） */
    public static final long EPOCH_SECOND = 1_704_067_200L;

    /** 内部盐值，用于校验位计算（可经 {@code -Ded.id.salt=...} 覆盖） */
    static final long CHECKSUM_SALT = resolveSalt();

    private static final int VERSION_BITS = 3;
    private static final int TS_BITS = 29;
    private static final int WORKER_BITS = 7;
    private static final int SEQ_BITS = 9;
    private static final int CHK_BITS = 4;

    private static final int CHK_SHIFT = 0;
    private static final int SEQ_SHIFT = CHK_BITS;
    private static final int WORKER_SHIFT = SEQ_SHIFT + SEQ_BITS;
    private static final int TS_SHIFT = WORKER_SHIFT + WORKER_BITS;
    private static final int VER_SHIFT = TS_SHIFT + TS_BITS;

    private static final int CHK_MASK = (1 << CHK_BITS) - 1;
    private static final int SEQ_MASK = (1 << SEQ_BITS) - 1;
    private static final int WORKER_MASK = (1 << WORKER_BITS) - 1;
    private static final long TS_MASK = (1L << TS_BITS) - 1;
    private static final int VER_MASK = (1 << VERSION_BITS) - 1;

    /** 本格式下理论最大 ID（52 位） */
    public static final long MAX_ID = (1L << (VER_SHIFT + VERSION_BITS)) - 1;

    private VerifiableOrderedId() {
    }

    /**
     * 组装 ID（由 {@link VerifiableOrderedIdGenerator} 调用）。
     */
    static long compose(int version, long epochSecond, int workerId, int sequence) {
        long relativeSecond = epochSecond - EPOCH_SECOND;
        if (relativeSecond < 0 || relativeSecond > TS_MASK) {
            throw new IllegalStateException("VOID 时间戳超出可编码范围: " + epochSecond);
        }
        int checksum = computeChecksum(version, relativeSecond, workerId, sequence);
        return ((long) (version & VER_MASK) << VER_SHIFT)
                | ((relativeSecond & TS_MASK) << TS_SHIFT)
                | ((long) (workerId & WORKER_MASK) << WORKER_SHIFT)
                | ((long) (sequence & SEQ_MASK) << SEQ_SHIFT)
                | (checksum & CHK_MASK);
    }

    /**
     * 判断 ID 是否符合 VOID 格式（版本、时间窗口、校验位）。
     */
    public static boolean isValid(long id) {
        return validate(id, 60L, null).valid();
    }

    /**
     * 带时钟容差与最老时间的校验，供拦截器使用。
     *
     * @param id               待校验 ID
     * @param futureToleranceSec 允许相对当前时间向未来的秒数
     * @param maxAgeSec        若为 null 则不限制最老时间；否则拒绝早于「当前 - maxAge」的 ID
     */
    public static ValidationResult validate(long id, long futureToleranceSec, Long maxAgeSec) {
        if (id <= 0 || id > MAX_ID) {
            return ValidationResult.invalid("ID 超出 VOID 数值范围");
        }
        int version = extractVersion(id);
        if (version != CURRENT_VERSION) {
            return ValidationResult.invalid("ID 版本不匹配");
        }
        long absoluteSecond = EPOCH_SECOND + extractRelativeSecond(id);
        long now = Instant.now().getEpochSecond();
        if (absoluteSecond > now + futureToleranceSec) {
            return ValidationResult.invalid("ID 时间戳位于未来");
        }
        if (maxAgeSec != null && absoluteSecond < now - maxAgeSec) {
            return ValidationResult.invalid("ID 时间戳过旧");
        }
        int worker = extractWorker(id);
        int sequence = extractSequence(id);
        int checksum = extractChecksum(id);
        if (checksum != computeChecksum(version, extractRelativeSecond(id), worker, sequence)) {
            return ValidationResult.invalid("ID 校验位不匹配");
        }
        return ValidationResult.ok(new Components(version, absoluteSecond, worker, sequence, checksum));
    }

    /**
     * 解析 ID 组成部分；非法时返回 empty。
     */
    public static Optional<Components> parse(long id) {
        ValidationResult result = validate(id, 60L, null);
        return result.valid() ? Optional.of(result.components()) : Optional.empty();
    }

    static int computeChecksum(int version, long relativeSecond, int workerId, int sequence) {
        long mix = (long) version * 0x9E3779B97F4A7C15L
                ^ relativeSecond * 0xC6A4A7935BD1E995L
                ^ (long) workerId * 0x165667B19E3779F9L
                ^ (long) sequence * 0x85EBCA77C2B2AE63L
                ^ CHECKSUM_SALT;
        return (int) ((mix ^ (mix >>> 32)) & CHK_MASK);
    }

    static int extractVersion(long id) {
        return (int) ((id >>> VER_SHIFT) & VER_MASK);
    }

    static long extractRelativeSecond(long id) {
        return (id >>> TS_SHIFT) & TS_MASK;
    }

    static int extractWorker(long id) {
        return (int) ((id >>> WORKER_SHIFT) & WORKER_MASK);
    }

    static int extractSequence(long id) {
        return (int) ((id >>> SEQ_SHIFT) & SEQ_MASK);
    }

    static int extractChecksum(long id) {
        return (int) (id & CHK_MASK);
    }

    private static long resolveSalt() {
        String configured = System.getProperty("ed.id.salt");
        if (configured != null && !configured.isBlank()) {
            try {
                return Long.parseUnsignedLong(configured.trim());
            } catch (NumberFormatException ignored) {
                return configured.trim().hashCode() & 0xFFFFFFFFL;
            }
        }
        // 默认值：与项目代号相关，生产环境建议 -Ded.id.salt 覆盖
        return 0x5A494544_2024L;
    }

    public record Components(int version, long epochSecond, int workerId, int sequence, int checksum) {
    }

    public record ValidationResult(boolean valid, String reason, Components components) {
        static ValidationResult ok(Components components) {
            return new ValidationResult(true, null, components);
        }

        static ValidationResult invalid(String reason) {
            return new ValidationResult(false, reason, null);
        }
    }
}
