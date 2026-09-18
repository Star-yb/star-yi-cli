package com.star.config.jimmer;

import org.babyfish.jimmer.sql.meta.DatabaseNamingStrategy;
import org.babyfish.jimmer.sql.runtime.DefaultDatabaseNamingStrategy;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration

public class JimmerConfigure {

    @Bean
    public DatabaseNamingStrategy databaseNamingStrategy() {
        // 默认命名策略 UPPER_CASE 大写  LOWER_CASE 小写
        return DefaultDatabaseNamingStrategy.LOWER_CASE;
    }
}
