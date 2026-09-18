package com.star.common.satoken;

import org.babyfish.jimmer.sql.runtime.AbstractScalarProvider;
import org.springframework.stereotype.Component;

import java.util.UUID;

@Component
public class UUIDScalarProvider extends AbstractScalarProvider<UUID, String> {

    @Override
    public UUID toScalar(String sqlValue) {
        return UUID.fromString(sqlValue);
    }

    @Override
    public String toSql(UUID scalarValue) {
        return scalarValue.toString();
    }


}