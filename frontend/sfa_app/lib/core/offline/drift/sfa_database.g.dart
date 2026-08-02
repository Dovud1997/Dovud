// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'sfa_database.dart';

// ignore_for_file: type=lint
class $CachedEntitiesTable extends CachedEntities
    with TableInfo<$CachedEntitiesTable, CachedEntityRow> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $CachedEntitiesTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _entityTypeMeta =
      const VerificationMeta('entityType');
  @override
  late final GeneratedColumn<String> entityType = GeneratedColumn<String>(
      'entity_type', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _entityIdMeta =
      const VerificationMeta('entityId');
  @override
  late final GeneratedColumn<String> entityId = GeneratedColumn<String>(
      'entity_id', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _payloadJsonMeta =
      const VerificationMeta('payloadJson');
  @override
  late final GeneratedColumn<String> payloadJson = GeneratedColumn<String>(
      'payload_json', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _updatedAtMeta =
      const VerificationMeta('updatedAt');
  @override
  late final GeneratedColumn<int> updatedAt = GeneratedColumn<int>(
      'updated_at', aliasedName, false,
      type: DriftSqlType.int, requiredDuringInsert: true);
  @override
  List<GeneratedColumn> get $columns =>
      [entityType, entityId, payloadJson, updatedAt];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'cached_entities';
  @override
  VerificationContext validateIntegrity(Insertable<CachedEntityRow> instance,
      {bool isInserting = false}) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('entity_type')) {
      context.handle(
          _entityTypeMeta,
          entityType.isAcceptableOrUnknown(
              data['entity_type']!, _entityTypeMeta));
    } else if (isInserting) {
      context.missing(_entityTypeMeta);
    }
    if (data.containsKey('entity_id')) {
      context.handle(_entityIdMeta,
          entityId.isAcceptableOrUnknown(data['entity_id']!, _entityIdMeta));
    } else if (isInserting) {
      context.missing(_entityIdMeta);
    }
    if (data.containsKey('payload_json')) {
      context.handle(
          _payloadJsonMeta,
          payloadJson.isAcceptableOrUnknown(
              data['payload_json']!, _payloadJsonMeta));
    } else if (isInserting) {
      context.missing(_payloadJsonMeta);
    }
    if (data.containsKey('updated_at')) {
      context.handle(_updatedAtMeta,
          updatedAt.isAcceptableOrUnknown(data['updated_at']!, _updatedAtMeta));
    } else if (isInserting) {
      context.missing(_updatedAtMeta);
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {entityType, entityId};
  @override
  CachedEntityRow map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return CachedEntityRow(
      entityType: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}entity_type'])!,
      entityId: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}entity_id'])!,
      payloadJson: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}payload_json'])!,
      updatedAt: attachedDatabase.typeMapping
          .read(DriftSqlType.int, data['${effectivePrefix}updated_at'])!,
    );
  }

  @override
  $CachedEntitiesTable createAlias(String alias) {
    return $CachedEntitiesTable(attachedDatabase, alias);
  }
}

class CachedEntityRow extends DataClass implements Insertable<CachedEntityRow> {
  final String entityType;
  final String entityId;
  final String payloadJson;
  final int updatedAt;
  const CachedEntityRow(
      {required this.entityType,
      required this.entityId,
      required this.payloadJson,
      required this.updatedAt});
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['entity_type'] = Variable<String>(entityType);
    map['entity_id'] = Variable<String>(entityId);
    map['payload_json'] = Variable<String>(payloadJson);
    map['updated_at'] = Variable<int>(updatedAt);
    return map;
  }

  CachedEntitiesCompanion toCompanion(bool nullToAbsent) {
    return CachedEntitiesCompanion(
      entityType: Value(entityType),
      entityId: Value(entityId),
      payloadJson: Value(payloadJson),
      updatedAt: Value(updatedAt),
    );
  }

  factory CachedEntityRow.fromJson(Map<String, dynamic> json,
      {ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return CachedEntityRow(
      entityType: serializer.fromJson<String>(json['entityType']),
      entityId: serializer.fromJson<String>(json['entityId']),
      payloadJson: serializer.fromJson<String>(json['payloadJson']),
      updatedAt: serializer.fromJson<int>(json['updatedAt']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'entityType': serializer.toJson<String>(entityType),
      'entityId': serializer.toJson<String>(entityId),
      'payloadJson': serializer.toJson<String>(payloadJson),
      'updatedAt': serializer.toJson<int>(updatedAt),
    };
  }

  CachedEntityRow copyWith(
          {String? entityType,
          String? entityId,
          String? payloadJson,
          int? updatedAt}) =>
      CachedEntityRow(
        entityType: entityType ?? this.entityType,
        entityId: entityId ?? this.entityId,
        payloadJson: payloadJson ?? this.payloadJson,
        updatedAt: updatedAt ?? this.updatedAt,
      );
  CachedEntityRow copyWithCompanion(CachedEntitiesCompanion data) {
    return CachedEntityRow(
      entityType:
          data.entityType.present ? data.entityType.value : this.entityType,
      entityId: data.entityId.present ? data.entityId.value : this.entityId,
      payloadJson:
          data.payloadJson.present ? data.payloadJson.value : this.payloadJson,
      updatedAt: data.updatedAt.present ? data.updatedAt.value : this.updatedAt,
    );
  }

  @override
  String toString() {
    return (StringBuffer('CachedEntityRow(')
          ..write('entityType: $entityType, ')
          ..write('entityId: $entityId, ')
          ..write('payloadJson: $payloadJson, ')
          ..write('updatedAt: $updatedAt')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(entityType, entityId, payloadJson, updatedAt);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is CachedEntityRow &&
          other.entityType == this.entityType &&
          other.entityId == this.entityId &&
          other.payloadJson == this.payloadJson &&
          other.updatedAt == this.updatedAt);
}

class CachedEntitiesCompanion extends UpdateCompanion<CachedEntityRow> {
  final Value<String> entityType;
  final Value<String> entityId;
  final Value<String> payloadJson;
  final Value<int> updatedAt;
  final Value<int> rowid;
  const CachedEntitiesCompanion({
    this.entityType = const Value.absent(),
    this.entityId = const Value.absent(),
    this.payloadJson = const Value.absent(),
    this.updatedAt = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  CachedEntitiesCompanion.insert({
    required String entityType,
    required String entityId,
    required String payloadJson,
    required int updatedAt,
    this.rowid = const Value.absent(),
  })  : entityType = Value(entityType),
        entityId = Value(entityId),
        payloadJson = Value(payloadJson),
        updatedAt = Value(updatedAt);
  static Insertable<CachedEntityRow> custom({
    Expression<String>? entityType,
    Expression<String>? entityId,
    Expression<String>? payloadJson,
    Expression<int>? updatedAt,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (entityType != null) 'entity_type': entityType,
      if (entityId != null) 'entity_id': entityId,
      if (payloadJson != null) 'payload_json': payloadJson,
      if (updatedAt != null) 'updated_at': updatedAt,
      if (rowid != null) 'rowid': rowid,
    });
  }

  CachedEntitiesCompanion copyWith(
      {Value<String>? entityType,
      Value<String>? entityId,
      Value<String>? payloadJson,
      Value<int>? updatedAt,
      Value<int>? rowid}) {
    return CachedEntitiesCompanion(
      entityType: entityType ?? this.entityType,
      entityId: entityId ?? this.entityId,
      payloadJson: payloadJson ?? this.payloadJson,
      updatedAt: updatedAt ?? this.updatedAt,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (entityType.present) {
      map['entity_type'] = Variable<String>(entityType.value);
    }
    if (entityId.present) {
      map['entity_id'] = Variable<String>(entityId.value);
    }
    if (payloadJson.present) {
      map['payload_json'] = Variable<String>(payloadJson.value);
    }
    if (updatedAt.present) {
      map['updated_at'] = Variable<int>(updatedAt.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('CachedEntitiesCompanion(')
          ..write('entityType: $entityType, ')
          ..write('entityId: $entityId, ')
          ..write('payloadJson: $payloadJson, ')
          ..write('updatedAt: $updatedAt, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

class $SyncMetaTable extends SyncMeta
    with TableInfo<$SyncMetaTable, SyncMetaData> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $SyncMetaTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _keyMeta = const VerificationMeta('key');
  @override
  late final GeneratedColumn<String> key = GeneratedColumn<String>(
      'key', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _valueMeta = const VerificationMeta('value');
  @override
  late final GeneratedColumn<String> value = GeneratedColumn<String>(
      'value', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  @override
  List<GeneratedColumn> get $columns => [key, value];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'sync_meta';
  @override
  VerificationContext validateIntegrity(Insertable<SyncMetaData> instance,
      {bool isInserting = false}) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('key')) {
      context.handle(
          _keyMeta, key.isAcceptableOrUnknown(data['key']!, _keyMeta));
    } else if (isInserting) {
      context.missing(_keyMeta);
    }
    if (data.containsKey('value')) {
      context.handle(
          _valueMeta, value.isAcceptableOrUnknown(data['value']!, _valueMeta));
    } else if (isInserting) {
      context.missing(_valueMeta);
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {key};
  @override
  SyncMetaData map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return SyncMetaData(
      key: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}key'])!,
      value: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}value'])!,
    );
  }

  @override
  $SyncMetaTable createAlias(String alias) {
    return $SyncMetaTable(attachedDatabase, alias);
  }
}

class SyncMetaData extends DataClass implements Insertable<SyncMetaData> {
  final String key;
  final String value;
  const SyncMetaData({required this.key, required this.value});
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['key'] = Variable<String>(key);
    map['value'] = Variable<String>(value);
    return map;
  }

  SyncMetaCompanion toCompanion(bool nullToAbsent) {
    return SyncMetaCompanion(
      key: Value(key),
      value: Value(value),
    );
  }

  factory SyncMetaData.fromJson(Map<String, dynamic> json,
      {ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return SyncMetaData(
      key: serializer.fromJson<String>(json['key']),
      value: serializer.fromJson<String>(json['value']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'key': serializer.toJson<String>(key),
      'value': serializer.toJson<String>(value),
    };
  }

  SyncMetaData copyWith({String? key, String? value}) => SyncMetaData(
        key: key ?? this.key,
        value: value ?? this.value,
      );
  SyncMetaData copyWithCompanion(SyncMetaCompanion data) {
    return SyncMetaData(
      key: data.key.present ? data.key.value : this.key,
      value: data.value.present ? data.value.value : this.value,
    );
  }

  @override
  String toString() {
    return (StringBuffer('SyncMetaData(')
          ..write('key: $key, ')
          ..write('value: $value')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(key, value);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is SyncMetaData &&
          other.key == this.key &&
          other.value == this.value);
}

class SyncMetaCompanion extends UpdateCompanion<SyncMetaData> {
  final Value<String> key;
  final Value<String> value;
  final Value<int> rowid;
  const SyncMetaCompanion({
    this.key = const Value.absent(),
    this.value = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  SyncMetaCompanion.insert({
    required String key,
    required String value,
    this.rowid = const Value.absent(),
  })  : key = Value(key),
        value = Value(value);
  static Insertable<SyncMetaData> custom({
    Expression<String>? key,
    Expression<String>? value,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (key != null) 'key': key,
      if (value != null) 'value': value,
      if (rowid != null) 'rowid': rowid,
    });
  }

  SyncMetaCompanion copyWith(
      {Value<String>? key, Value<String>? value, Value<int>? rowid}) {
    return SyncMetaCompanion(
      key: key ?? this.key,
      value: value ?? this.value,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (key.present) {
      map['key'] = Variable<String>(key.value);
    }
    if (value.present) {
      map['value'] = Variable<String>(value.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('SyncMetaCompanion(')
          ..write('key: $key, ')
          ..write('value: $value, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

class $OutboxOpsTable extends OutboxOps
    with TableInfo<$OutboxOpsTable, OutboxOpRow> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $OutboxOpsTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _opIdMeta = const VerificationMeta('opId');
  @override
  late final GeneratedColumn<String> opId = GeneratedColumn<String>(
      'op_id', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _entityTypeMeta =
      const VerificationMeta('entityType');
  @override
  late final GeneratedColumn<String> entityType = GeneratedColumn<String>(
      'entity_type', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _entityIdMeta =
      const VerificationMeta('entityId');
  @override
  late final GeneratedColumn<String> entityId = GeneratedColumn<String>(
      'entity_id', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _opMeta = const VerificationMeta('op');
  @override
  late final GeneratedColumn<String> op = GeneratedColumn<String>(
      'op', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _baseVersionMeta =
      const VerificationMeta('baseVersion');
  @override
  late final GeneratedColumn<int> baseVersion = GeneratedColumn<int>(
      'base_version', aliasedName, false,
      type: DriftSqlType.int,
      requiredDuringInsert: false,
      defaultValue: const Constant(0));
  static const VerificationMeta _payloadJsonMeta =
      const VerificationMeta('payloadJson');
  @override
  late final GeneratedColumn<String> payloadJson = GeneratedColumn<String>(
      'payload_json', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _statusMeta = const VerificationMeta('status');
  @override
  late final GeneratedColumn<String> status = GeneratedColumn<String>(
      'status', aliasedName, false,
      type: DriftSqlType.string,
      requiredDuringInsert: false,
      defaultValue: const Constant('pending'));
  static const VerificationMeta _createdAtMeta =
      const VerificationMeta('createdAt');
  @override
  late final GeneratedColumn<DateTime> createdAt = GeneratedColumn<DateTime>(
      'created_at', aliasedName, false,
      type: DriftSqlType.dateTime, requiredDuringInsert: true);
  @override
  List<GeneratedColumn> get $columns => [
        opId,
        entityType,
        entityId,
        op,
        baseVersion,
        payloadJson,
        status,
        createdAt
      ];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'outbox_ops';
  @override
  VerificationContext validateIntegrity(Insertable<OutboxOpRow> instance,
      {bool isInserting = false}) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('op_id')) {
      context.handle(
          _opIdMeta, opId.isAcceptableOrUnknown(data['op_id']!, _opIdMeta));
    } else if (isInserting) {
      context.missing(_opIdMeta);
    }
    if (data.containsKey('entity_type')) {
      context.handle(
          _entityTypeMeta,
          entityType.isAcceptableOrUnknown(
              data['entity_type']!, _entityTypeMeta));
    } else if (isInserting) {
      context.missing(_entityTypeMeta);
    }
    if (data.containsKey('entity_id')) {
      context.handle(_entityIdMeta,
          entityId.isAcceptableOrUnknown(data['entity_id']!, _entityIdMeta));
    } else if (isInserting) {
      context.missing(_entityIdMeta);
    }
    if (data.containsKey('op')) {
      context.handle(_opMeta, op.isAcceptableOrUnknown(data['op']!, _opMeta));
    } else if (isInserting) {
      context.missing(_opMeta);
    }
    if (data.containsKey('base_version')) {
      context.handle(
          _baseVersionMeta,
          baseVersion.isAcceptableOrUnknown(
              data['base_version']!, _baseVersionMeta));
    }
    if (data.containsKey('payload_json')) {
      context.handle(
          _payloadJsonMeta,
          payloadJson.isAcceptableOrUnknown(
              data['payload_json']!, _payloadJsonMeta));
    } else if (isInserting) {
      context.missing(_payloadJsonMeta);
    }
    if (data.containsKey('status')) {
      context.handle(_statusMeta,
          status.isAcceptableOrUnknown(data['status']!, _statusMeta));
    }
    if (data.containsKey('created_at')) {
      context.handle(_createdAtMeta,
          createdAt.isAcceptableOrUnknown(data['created_at']!, _createdAtMeta));
    } else if (isInserting) {
      context.missing(_createdAtMeta);
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {opId};
  @override
  OutboxOpRow map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return OutboxOpRow(
      opId: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}op_id'])!,
      entityType: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}entity_type'])!,
      entityId: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}entity_id'])!,
      op: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}op'])!,
      baseVersion: attachedDatabase.typeMapping
          .read(DriftSqlType.int, data['${effectivePrefix}base_version'])!,
      payloadJson: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}payload_json'])!,
      status: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}status'])!,
      createdAt: attachedDatabase.typeMapping
          .read(DriftSqlType.dateTime, data['${effectivePrefix}created_at'])!,
    );
  }

  @override
  $OutboxOpsTable createAlias(String alias) {
    return $OutboxOpsTable(attachedDatabase, alias);
  }
}

class OutboxOpRow extends DataClass implements Insertable<OutboxOpRow> {
  final String opId;
  final String entityType;
  final String entityId;
  final String op;
  final int baseVersion;
  final String payloadJson;
  final String status;
  final DateTime createdAt;
  const OutboxOpRow(
      {required this.opId,
      required this.entityType,
      required this.entityId,
      required this.op,
      required this.baseVersion,
      required this.payloadJson,
      required this.status,
      required this.createdAt});
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['op_id'] = Variable<String>(opId);
    map['entity_type'] = Variable<String>(entityType);
    map['entity_id'] = Variable<String>(entityId);
    map['op'] = Variable<String>(op);
    map['base_version'] = Variable<int>(baseVersion);
    map['payload_json'] = Variable<String>(payloadJson);
    map['status'] = Variable<String>(status);
    map['created_at'] = Variable<DateTime>(createdAt);
    return map;
  }

  OutboxOpsCompanion toCompanion(bool nullToAbsent) {
    return OutboxOpsCompanion(
      opId: Value(opId),
      entityType: Value(entityType),
      entityId: Value(entityId),
      op: Value(op),
      baseVersion: Value(baseVersion),
      payloadJson: Value(payloadJson),
      status: Value(status),
      createdAt: Value(createdAt),
    );
  }

  factory OutboxOpRow.fromJson(Map<String, dynamic> json,
      {ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return OutboxOpRow(
      opId: serializer.fromJson<String>(json['opId']),
      entityType: serializer.fromJson<String>(json['entityType']),
      entityId: serializer.fromJson<String>(json['entityId']),
      op: serializer.fromJson<String>(json['op']),
      baseVersion: serializer.fromJson<int>(json['baseVersion']),
      payloadJson: serializer.fromJson<String>(json['payloadJson']),
      status: serializer.fromJson<String>(json['status']),
      createdAt: serializer.fromJson<DateTime>(json['createdAt']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'opId': serializer.toJson<String>(opId),
      'entityType': serializer.toJson<String>(entityType),
      'entityId': serializer.toJson<String>(entityId),
      'op': serializer.toJson<String>(op),
      'baseVersion': serializer.toJson<int>(baseVersion),
      'payloadJson': serializer.toJson<String>(payloadJson),
      'status': serializer.toJson<String>(status),
      'createdAt': serializer.toJson<DateTime>(createdAt),
    };
  }

  OutboxOpRow copyWith(
          {String? opId,
          String? entityType,
          String? entityId,
          String? op,
          int? baseVersion,
          String? payloadJson,
          String? status,
          DateTime? createdAt}) =>
      OutboxOpRow(
        opId: opId ?? this.opId,
        entityType: entityType ?? this.entityType,
        entityId: entityId ?? this.entityId,
        op: op ?? this.op,
        baseVersion: baseVersion ?? this.baseVersion,
        payloadJson: payloadJson ?? this.payloadJson,
        status: status ?? this.status,
        createdAt: createdAt ?? this.createdAt,
      );
  OutboxOpRow copyWithCompanion(OutboxOpsCompanion data) {
    return OutboxOpRow(
      opId: data.opId.present ? data.opId.value : this.opId,
      entityType:
          data.entityType.present ? data.entityType.value : this.entityType,
      entityId: data.entityId.present ? data.entityId.value : this.entityId,
      op: data.op.present ? data.op.value : this.op,
      baseVersion:
          data.baseVersion.present ? data.baseVersion.value : this.baseVersion,
      payloadJson:
          data.payloadJson.present ? data.payloadJson.value : this.payloadJson,
      status: data.status.present ? data.status.value : this.status,
      createdAt: data.createdAt.present ? data.createdAt.value : this.createdAt,
    );
  }

  @override
  String toString() {
    return (StringBuffer('OutboxOpRow(')
          ..write('opId: $opId, ')
          ..write('entityType: $entityType, ')
          ..write('entityId: $entityId, ')
          ..write('op: $op, ')
          ..write('baseVersion: $baseVersion, ')
          ..write('payloadJson: $payloadJson, ')
          ..write('status: $status, ')
          ..write('createdAt: $createdAt')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(opId, entityType, entityId, op, baseVersion,
      payloadJson, status, createdAt);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is OutboxOpRow &&
          other.opId == this.opId &&
          other.entityType == this.entityType &&
          other.entityId == this.entityId &&
          other.op == this.op &&
          other.baseVersion == this.baseVersion &&
          other.payloadJson == this.payloadJson &&
          other.status == this.status &&
          other.createdAt == this.createdAt);
}

class OutboxOpsCompanion extends UpdateCompanion<OutboxOpRow> {
  final Value<String> opId;
  final Value<String> entityType;
  final Value<String> entityId;
  final Value<String> op;
  final Value<int> baseVersion;
  final Value<String> payloadJson;
  final Value<String> status;
  final Value<DateTime> createdAt;
  final Value<int> rowid;
  const OutboxOpsCompanion({
    this.opId = const Value.absent(),
    this.entityType = const Value.absent(),
    this.entityId = const Value.absent(),
    this.op = const Value.absent(),
    this.baseVersion = const Value.absent(),
    this.payloadJson = const Value.absent(),
    this.status = const Value.absent(),
    this.createdAt = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  OutboxOpsCompanion.insert({
    required String opId,
    required String entityType,
    required String entityId,
    required String op,
    this.baseVersion = const Value.absent(),
    required String payloadJson,
    this.status = const Value.absent(),
    required DateTime createdAt,
    this.rowid = const Value.absent(),
  })  : opId = Value(opId),
        entityType = Value(entityType),
        entityId = Value(entityId),
        op = Value(op),
        payloadJson = Value(payloadJson),
        createdAt = Value(createdAt);
  static Insertable<OutboxOpRow> custom({
    Expression<String>? opId,
    Expression<String>? entityType,
    Expression<String>? entityId,
    Expression<String>? op,
    Expression<int>? baseVersion,
    Expression<String>? payloadJson,
    Expression<String>? status,
    Expression<DateTime>? createdAt,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (opId != null) 'op_id': opId,
      if (entityType != null) 'entity_type': entityType,
      if (entityId != null) 'entity_id': entityId,
      if (op != null) 'op': op,
      if (baseVersion != null) 'base_version': baseVersion,
      if (payloadJson != null) 'payload_json': payloadJson,
      if (status != null) 'status': status,
      if (createdAt != null) 'created_at': createdAt,
      if (rowid != null) 'rowid': rowid,
    });
  }

  OutboxOpsCompanion copyWith(
      {Value<String>? opId,
      Value<String>? entityType,
      Value<String>? entityId,
      Value<String>? op,
      Value<int>? baseVersion,
      Value<String>? payloadJson,
      Value<String>? status,
      Value<DateTime>? createdAt,
      Value<int>? rowid}) {
    return OutboxOpsCompanion(
      opId: opId ?? this.opId,
      entityType: entityType ?? this.entityType,
      entityId: entityId ?? this.entityId,
      op: op ?? this.op,
      baseVersion: baseVersion ?? this.baseVersion,
      payloadJson: payloadJson ?? this.payloadJson,
      status: status ?? this.status,
      createdAt: createdAt ?? this.createdAt,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (opId.present) {
      map['op_id'] = Variable<String>(opId.value);
    }
    if (entityType.present) {
      map['entity_type'] = Variable<String>(entityType.value);
    }
    if (entityId.present) {
      map['entity_id'] = Variable<String>(entityId.value);
    }
    if (op.present) {
      map['op'] = Variable<String>(op.value);
    }
    if (baseVersion.present) {
      map['base_version'] = Variable<int>(baseVersion.value);
    }
    if (payloadJson.present) {
      map['payload_json'] = Variable<String>(payloadJson.value);
    }
    if (status.present) {
      map['status'] = Variable<String>(status.value);
    }
    if (createdAt.present) {
      map['created_at'] = Variable<DateTime>(createdAt.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('OutboxOpsCompanion(')
          ..write('opId: $opId, ')
          ..write('entityType: $entityType, ')
          ..write('entityId: $entityId, ')
          ..write('op: $op, ')
          ..write('baseVersion: $baseVersion, ')
          ..write('payloadJson: $payloadJson, ')
          ..write('status: $status, ')
          ..write('createdAt: $createdAt, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

class $FileUploadsTable extends FileUploads
    with TableInfo<$FileUploadsTable, FileUploadRow> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $FileUploadsTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _uploadIdMeta =
      const VerificationMeta('uploadId');
  @override
  late final GeneratedColumn<String> uploadId = GeneratedColumn<String>(
      'upload_id', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _fileNameMeta =
      const VerificationMeta('fileName');
  @override
  late final GeneratedColumn<String> fileName = GeneratedColumn<String>(
      'file_name', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _mimeMeta = const VerificationMeta('mime');
  @override
  late final GeneratedColumn<String> mime = GeneratedColumn<String>(
      'mime', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _sizeBytesMeta =
      const VerificationMeta('sizeBytes');
  @override
  late final GeneratedColumn<int> sizeBytes = GeneratedColumn<int>(
      'size_bytes', aliasedName, false,
      type: DriftSqlType.int,
      requiredDuringInsert: false,
      defaultValue: const Constant(0));
  static const VerificationMeta _localPathMeta =
      const VerificationMeta('localPath');
  @override
  late final GeneratedColumn<String> localPath = GeneratedColumn<String>(
      'local_path', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _statusMeta = const VerificationMeta('status');
  @override
  late final GeneratedColumn<String> status = GeneratedColumn<String>(
      'status', aliasedName, false,
      type: DriftSqlType.string,
      requiredDuringInsert: false,
      defaultValue: const Constant('pending'));
  static const VerificationMeta _remoteFileIdMeta =
      const VerificationMeta('remoteFileId');
  @override
  late final GeneratedColumn<String> remoteFileId = GeneratedColumn<String>(
      'remote_file_id', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _errorMeta = const VerificationMeta('error');
  @override
  late final GeneratedColumn<String> error = GeneratedColumn<String>(
      'error', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _createdAtMeta =
      const VerificationMeta('createdAt');
  @override
  late final GeneratedColumn<DateTime> createdAt = GeneratedColumn<DateTime>(
      'created_at', aliasedName, false,
      type: DriftSqlType.dateTime, requiredDuringInsert: true);
  static const VerificationMeta _payloadMeta =
      const VerificationMeta('payload');
  @override
  late final GeneratedColumn<Uint8List> payload = GeneratedColumn<Uint8List>(
      'payload', aliasedName, true,
      type: DriftSqlType.blob, requiredDuringInsert: false);
  @override
  List<GeneratedColumn> get $columns => [
        uploadId,
        fileName,
        mime,
        sizeBytes,
        localPath,
        status,
        remoteFileId,
        error,
        createdAt,
        payload
      ];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'file_uploads';
  @override
  VerificationContext validateIntegrity(Insertable<FileUploadRow> instance,
      {bool isInserting = false}) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('upload_id')) {
      context.handle(_uploadIdMeta,
          uploadId.isAcceptableOrUnknown(data['upload_id']!, _uploadIdMeta));
    } else if (isInserting) {
      context.missing(_uploadIdMeta);
    }
    if (data.containsKey('file_name')) {
      context.handle(_fileNameMeta,
          fileName.isAcceptableOrUnknown(data['file_name']!, _fileNameMeta));
    } else if (isInserting) {
      context.missing(_fileNameMeta);
    }
    if (data.containsKey('mime')) {
      context.handle(
          _mimeMeta, mime.isAcceptableOrUnknown(data['mime']!, _mimeMeta));
    } else if (isInserting) {
      context.missing(_mimeMeta);
    }
    if (data.containsKey('size_bytes')) {
      context.handle(_sizeBytesMeta,
          sizeBytes.isAcceptableOrUnknown(data['size_bytes']!, _sizeBytesMeta));
    }
    if (data.containsKey('local_path')) {
      context.handle(_localPathMeta,
          localPath.isAcceptableOrUnknown(data['local_path']!, _localPathMeta));
    }
    if (data.containsKey('status')) {
      context.handle(_statusMeta,
          status.isAcceptableOrUnknown(data['status']!, _statusMeta));
    }
    if (data.containsKey('remote_file_id')) {
      context.handle(
          _remoteFileIdMeta,
          remoteFileId.isAcceptableOrUnknown(
              data['remote_file_id']!, _remoteFileIdMeta));
    }
    if (data.containsKey('error')) {
      context.handle(
          _errorMeta, error.isAcceptableOrUnknown(data['error']!, _errorMeta));
    }
    if (data.containsKey('created_at')) {
      context.handle(_createdAtMeta,
          createdAt.isAcceptableOrUnknown(data['created_at']!, _createdAtMeta));
    } else if (isInserting) {
      context.missing(_createdAtMeta);
    }
    if (data.containsKey('payload')) {
      context.handle(_payloadMeta,
          payload.isAcceptableOrUnknown(data['payload']!, _payloadMeta));
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {uploadId};
  @override
  FileUploadRow map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return FileUploadRow(
      uploadId: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}upload_id'])!,
      fileName: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}file_name'])!,
      mime: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}mime'])!,
      sizeBytes: attachedDatabase.typeMapping
          .read(DriftSqlType.int, data['${effectivePrefix}size_bytes'])!,
      localPath: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}local_path']),
      status: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}status'])!,
      remoteFileId: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}remote_file_id']),
      error: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}error']),
      createdAt: attachedDatabase.typeMapping
          .read(DriftSqlType.dateTime, data['${effectivePrefix}created_at'])!,
      payload: attachedDatabase.typeMapping
          .read(DriftSqlType.blob, data['${effectivePrefix}payload']),
    );
  }

  @override
  $FileUploadsTable createAlias(String alias) {
    return $FileUploadsTable(attachedDatabase, alias);
  }
}

class FileUploadRow extends DataClass implements Insertable<FileUploadRow> {
  final String uploadId;
  final String fileName;
  final String mime;
  final int sizeBytes;
  final String? localPath;
  final String status;
  final String? remoteFileId;
  final String? error;
  final DateTime createdAt;

  /// Raw file bytes so pending uploads survive process restarts.
  final Uint8List? payload;
  const FileUploadRow(
      {required this.uploadId,
      required this.fileName,
      required this.mime,
      required this.sizeBytes,
      this.localPath,
      required this.status,
      this.remoteFileId,
      this.error,
      required this.createdAt,
      this.payload});
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['upload_id'] = Variable<String>(uploadId);
    map['file_name'] = Variable<String>(fileName);
    map['mime'] = Variable<String>(mime);
    map['size_bytes'] = Variable<int>(sizeBytes);
    if (!nullToAbsent || localPath != null) {
      map['local_path'] = Variable<String>(localPath);
    }
    map['status'] = Variable<String>(status);
    if (!nullToAbsent || remoteFileId != null) {
      map['remote_file_id'] = Variable<String>(remoteFileId);
    }
    if (!nullToAbsent || error != null) {
      map['error'] = Variable<String>(error);
    }
    map['created_at'] = Variable<DateTime>(createdAt);
    if (!nullToAbsent || payload != null) {
      map['payload'] = Variable<Uint8List>(payload);
    }
    return map;
  }

  FileUploadsCompanion toCompanion(bool nullToAbsent) {
    return FileUploadsCompanion(
      uploadId: Value(uploadId),
      fileName: Value(fileName),
      mime: Value(mime),
      sizeBytes: Value(sizeBytes),
      localPath: localPath == null && nullToAbsent
          ? const Value.absent()
          : Value(localPath),
      status: Value(status),
      remoteFileId: remoteFileId == null && nullToAbsent
          ? const Value.absent()
          : Value(remoteFileId),
      error:
          error == null && nullToAbsent ? const Value.absent() : Value(error),
      createdAt: Value(createdAt),
      payload: payload == null && nullToAbsent
          ? const Value.absent()
          : Value(payload),
    );
  }

  factory FileUploadRow.fromJson(Map<String, dynamic> json,
      {ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return FileUploadRow(
      uploadId: serializer.fromJson<String>(json['uploadId']),
      fileName: serializer.fromJson<String>(json['fileName']),
      mime: serializer.fromJson<String>(json['mime']),
      sizeBytes: serializer.fromJson<int>(json['sizeBytes']),
      localPath: serializer.fromJson<String?>(json['localPath']),
      status: serializer.fromJson<String>(json['status']),
      remoteFileId: serializer.fromJson<String?>(json['remoteFileId']),
      error: serializer.fromJson<String?>(json['error']),
      createdAt: serializer.fromJson<DateTime>(json['createdAt']),
      payload: serializer.fromJson<Uint8List?>(json['payload']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'uploadId': serializer.toJson<String>(uploadId),
      'fileName': serializer.toJson<String>(fileName),
      'mime': serializer.toJson<String>(mime),
      'sizeBytes': serializer.toJson<int>(sizeBytes),
      'localPath': serializer.toJson<String?>(localPath),
      'status': serializer.toJson<String>(status),
      'remoteFileId': serializer.toJson<String?>(remoteFileId),
      'error': serializer.toJson<String?>(error),
      'createdAt': serializer.toJson<DateTime>(createdAt),
      'payload': serializer.toJson<Uint8List?>(payload),
    };
  }

  FileUploadRow copyWith(
          {String? uploadId,
          String? fileName,
          String? mime,
          int? sizeBytes,
          Value<String?> localPath = const Value.absent(),
          String? status,
          Value<String?> remoteFileId = const Value.absent(),
          Value<String?> error = const Value.absent(),
          DateTime? createdAt,
          Value<Uint8List?> payload = const Value.absent()}) =>
      FileUploadRow(
        uploadId: uploadId ?? this.uploadId,
        fileName: fileName ?? this.fileName,
        mime: mime ?? this.mime,
        sizeBytes: sizeBytes ?? this.sizeBytes,
        localPath: localPath.present ? localPath.value : this.localPath,
        status: status ?? this.status,
        remoteFileId:
            remoteFileId.present ? remoteFileId.value : this.remoteFileId,
        error: error.present ? error.value : this.error,
        createdAt: createdAt ?? this.createdAt,
        payload: payload.present ? payload.value : this.payload,
      );
  FileUploadRow copyWithCompanion(FileUploadsCompanion data) {
    return FileUploadRow(
      uploadId: data.uploadId.present ? data.uploadId.value : this.uploadId,
      fileName: data.fileName.present ? data.fileName.value : this.fileName,
      mime: data.mime.present ? data.mime.value : this.mime,
      sizeBytes: data.sizeBytes.present ? data.sizeBytes.value : this.sizeBytes,
      localPath: data.localPath.present ? data.localPath.value : this.localPath,
      status: data.status.present ? data.status.value : this.status,
      remoteFileId: data.remoteFileId.present
          ? data.remoteFileId.value
          : this.remoteFileId,
      error: data.error.present ? data.error.value : this.error,
      createdAt: data.createdAt.present ? data.createdAt.value : this.createdAt,
      payload: data.payload.present ? data.payload.value : this.payload,
    );
  }

  @override
  String toString() {
    return (StringBuffer('FileUploadRow(')
          ..write('uploadId: $uploadId, ')
          ..write('fileName: $fileName, ')
          ..write('mime: $mime, ')
          ..write('sizeBytes: $sizeBytes, ')
          ..write('localPath: $localPath, ')
          ..write('status: $status, ')
          ..write('remoteFileId: $remoteFileId, ')
          ..write('error: $error, ')
          ..write('createdAt: $createdAt, ')
          ..write('payload: $payload')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(
      uploadId,
      fileName,
      mime,
      sizeBytes,
      localPath,
      status,
      remoteFileId,
      error,
      createdAt,
      $driftBlobEquality.hash(payload));
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is FileUploadRow &&
          other.uploadId == this.uploadId &&
          other.fileName == this.fileName &&
          other.mime == this.mime &&
          other.sizeBytes == this.sizeBytes &&
          other.localPath == this.localPath &&
          other.status == this.status &&
          other.remoteFileId == this.remoteFileId &&
          other.error == this.error &&
          other.createdAt == this.createdAt &&
          $driftBlobEquality.equals(other.payload, this.payload));
}

class FileUploadsCompanion extends UpdateCompanion<FileUploadRow> {
  final Value<String> uploadId;
  final Value<String> fileName;
  final Value<String> mime;
  final Value<int> sizeBytes;
  final Value<String?> localPath;
  final Value<String> status;
  final Value<String?> remoteFileId;
  final Value<String?> error;
  final Value<DateTime> createdAt;
  final Value<Uint8List?> payload;
  final Value<int> rowid;
  const FileUploadsCompanion({
    this.uploadId = const Value.absent(),
    this.fileName = const Value.absent(),
    this.mime = const Value.absent(),
    this.sizeBytes = const Value.absent(),
    this.localPath = const Value.absent(),
    this.status = const Value.absent(),
    this.remoteFileId = const Value.absent(),
    this.error = const Value.absent(),
    this.createdAt = const Value.absent(),
    this.payload = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  FileUploadsCompanion.insert({
    required String uploadId,
    required String fileName,
    required String mime,
    this.sizeBytes = const Value.absent(),
    this.localPath = const Value.absent(),
    this.status = const Value.absent(),
    this.remoteFileId = const Value.absent(),
    this.error = const Value.absent(),
    required DateTime createdAt,
    this.payload = const Value.absent(),
    this.rowid = const Value.absent(),
  })  : uploadId = Value(uploadId),
        fileName = Value(fileName),
        mime = Value(mime),
        createdAt = Value(createdAt);
  static Insertable<FileUploadRow> custom({
    Expression<String>? uploadId,
    Expression<String>? fileName,
    Expression<String>? mime,
    Expression<int>? sizeBytes,
    Expression<String>? localPath,
    Expression<String>? status,
    Expression<String>? remoteFileId,
    Expression<String>? error,
    Expression<DateTime>? createdAt,
    Expression<Uint8List>? payload,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (uploadId != null) 'upload_id': uploadId,
      if (fileName != null) 'file_name': fileName,
      if (mime != null) 'mime': mime,
      if (sizeBytes != null) 'size_bytes': sizeBytes,
      if (localPath != null) 'local_path': localPath,
      if (status != null) 'status': status,
      if (remoteFileId != null) 'remote_file_id': remoteFileId,
      if (error != null) 'error': error,
      if (createdAt != null) 'created_at': createdAt,
      if (payload != null) 'payload': payload,
      if (rowid != null) 'rowid': rowid,
    });
  }

  FileUploadsCompanion copyWith(
      {Value<String>? uploadId,
      Value<String>? fileName,
      Value<String>? mime,
      Value<int>? sizeBytes,
      Value<String?>? localPath,
      Value<String>? status,
      Value<String?>? remoteFileId,
      Value<String?>? error,
      Value<DateTime>? createdAt,
      Value<Uint8List?>? payload,
      Value<int>? rowid}) {
    return FileUploadsCompanion(
      uploadId: uploadId ?? this.uploadId,
      fileName: fileName ?? this.fileName,
      mime: mime ?? this.mime,
      sizeBytes: sizeBytes ?? this.sizeBytes,
      localPath: localPath ?? this.localPath,
      status: status ?? this.status,
      remoteFileId: remoteFileId ?? this.remoteFileId,
      error: error ?? this.error,
      createdAt: createdAt ?? this.createdAt,
      payload: payload ?? this.payload,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (uploadId.present) {
      map['upload_id'] = Variable<String>(uploadId.value);
    }
    if (fileName.present) {
      map['file_name'] = Variable<String>(fileName.value);
    }
    if (mime.present) {
      map['mime'] = Variable<String>(mime.value);
    }
    if (sizeBytes.present) {
      map['size_bytes'] = Variable<int>(sizeBytes.value);
    }
    if (localPath.present) {
      map['local_path'] = Variable<String>(localPath.value);
    }
    if (status.present) {
      map['status'] = Variable<String>(status.value);
    }
    if (remoteFileId.present) {
      map['remote_file_id'] = Variable<String>(remoteFileId.value);
    }
    if (error.present) {
      map['error'] = Variable<String>(error.value);
    }
    if (createdAt.present) {
      map['created_at'] = Variable<DateTime>(createdAt.value);
    }
    if (payload.present) {
      map['payload'] = Variable<Uint8List>(payload.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('FileUploadsCompanion(')
          ..write('uploadId: $uploadId, ')
          ..write('fileName: $fileName, ')
          ..write('mime: $mime, ')
          ..write('sizeBytes: $sizeBytes, ')
          ..write('localPath: $localPath, ')
          ..write('status: $status, ')
          ..write('remoteFileId: $remoteFileId, ')
          ..write('error: $error, ')
          ..write('createdAt: $createdAt, ')
          ..write('payload: $payload, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

class $GpsPendingTable extends GpsPending
    with TableInfo<$GpsPendingTable, GpsPendingRow> {
  @override
  final GeneratedDatabase attachedDatabase;
  final String? _alias;
  $GpsPendingTable(this.attachedDatabase, [this._alias]);
  static const VerificationMeta _pointIdMeta =
      const VerificationMeta('pointId');
  @override
  late final GeneratedColumn<String> pointId = GeneratedColumn<String>(
      'point_id', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _agentIdMeta =
      const VerificationMeta('agentId');
  @override
  late final GeneratedColumn<String> agentId = GeneratedColumn<String>(
      'agent_id', aliasedName, false,
      type: DriftSqlType.string, requiredDuringInsert: true);
  static const VerificationMeta _visitIdMeta =
      const VerificationMeta('visitId');
  @override
  late final GeneratedColumn<String> visitId = GeneratedColumn<String>(
      'visit_id', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _latMeta = const VerificationMeta('lat');
  @override
  late final GeneratedColumn<double> lat = GeneratedColumn<double>(
      'lat', aliasedName, false,
      type: DriftSqlType.double, requiredDuringInsert: true);
  static const VerificationMeta _lngMeta = const VerificationMeta('lng');
  @override
  late final GeneratedColumn<double> lng = GeneratedColumn<double>(
      'lng', aliasedName, false,
      type: DriftSqlType.double, requiredDuringInsert: true);
  static const VerificationMeta _accuracyMeta =
      const VerificationMeta('accuracy');
  @override
  late final GeneratedColumn<double> accuracy = GeneratedColumn<double>(
      'accuracy', aliasedName, true,
      type: DriftSqlType.double, requiredDuringInsert: false);
  static const VerificationMeta _recordedAtMeta =
      const VerificationMeta('recordedAt');
  @override
  late final GeneratedColumn<DateTime> recordedAt = GeneratedColumn<DateTime>(
      'recorded_at', aliasedName, false,
      type: DriftSqlType.dateTime, requiredDuringInsert: true);
  static const VerificationMeta _statusMeta = const VerificationMeta('status');
  @override
  late final GeneratedColumn<String> status = GeneratedColumn<String>(
      'status', aliasedName, false,
      type: DriftSqlType.string,
      requiredDuringInsert: false,
      defaultValue: const Constant('pending'));
  static const VerificationMeta _errorMeta = const VerificationMeta('error');
  @override
  late final GeneratedColumn<String> error = GeneratedColumn<String>(
      'error', aliasedName, true,
      type: DriftSqlType.string, requiredDuringInsert: false);
  static const VerificationMeta _createdAtMeta =
      const VerificationMeta('createdAt');
  @override
  late final GeneratedColumn<DateTime> createdAt = GeneratedColumn<DateTime>(
      'created_at', aliasedName, false,
      type: DriftSqlType.dateTime, requiredDuringInsert: true);
  @override
  List<GeneratedColumn> get $columns => [
        pointId,
        agentId,
        visitId,
        lat,
        lng,
        accuracy,
        recordedAt,
        status,
        error,
        createdAt
      ];
  @override
  String get aliasedName => _alias ?? actualTableName;
  @override
  String get actualTableName => $name;
  static const String $name = 'gps_pending';
  @override
  VerificationContext validateIntegrity(Insertable<GpsPendingRow> instance,
      {bool isInserting = false}) {
    final context = VerificationContext();
    final data = instance.toColumns(true);
    if (data.containsKey('point_id')) {
      context.handle(_pointIdMeta,
          pointId.isAcceptableOrUnknown(data['point_id']!, _pointIdMeta));
    } else if (isInserting) {
      context.missing(_pointIdMeta);
    }
    if (data.containsKey('agent_id')) {
      context.handle(_agentIdMeta,
          agentId.isAcceptableOrUnknown(data['agent_id']!, _agentIdMeta));
    } else if (isInserting) {
      context.missing(_agentIdMeta);
    }
    if (data.containsKey('visit_id')) {
      context.handle(_visitIdMeta,
          visitId.isAcceptableOrUnknown(data['visit_id']!, _visitIdMeta));
    }
    if (data.containsKey('lat')) {
      context.handle(
          _latMeta, lat.isAcceptableOrUnknown(data['lat']!, _latMeta));
    } else if (isInserting) {
      context.missing(_latMeta);
    }
    if (data.containsKey('lng')) {
      context.handle(
          _lngMeta, lng.isAcceptableOrUnknown(data['lng']!, _lngMeta));
    } else if (isInserting) {
      context.missing(_lngMeta);
    }
    if (data.containsKey('accuracy')) {
      context.handle(_accuracyMeta,
          accuracy.isAcceptableOrUnknown(data['accuracy']!, _accuracyMeta));
    }
    if (data.containsKey('recorded_at')) {
      context.handle(
          _recordedAtMeta,
          recordedAt.isAcceptableOrUnknown(
              data['recorded_at']!, _recordedAtMeta));
    } else if (isInserting) {
      context.missing(_recordedAtMeta);
    }
    if (data.containsKey('status')) {
      context.handle(_statusMeta,
          status.isAcceptableOrUnknown(data['status']!, _statusMeta));
    }
    if (data.containsKey('error')) {
      context.handle(
          _errorMeta, error.isAcceptableOrUnknown(data['error']!, _errorMeta));
    }
    if (data.containsKey('created_at')) {
      context.handle(_createdAtMeta,
          createdAt.isAcceptableOrUnknown(data['created_at']!, _createdAtMeta));
    } else if (isInserting) {
      context.missing(_createdAtMeta);
    }
    return context;
  }

  @override
  Set<GeneratedColumn> get $primaryKey => {pointId};
  @override
  GpsPendingRow map(Map<String, dynamic> data, {String? tablePrefix}) {
    final effectivePrefix = tablePrefix != null ? '$tablePrefix.' : '';
    return GpsPendingRow(
      pointId: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}point_id'])!,
      agentId: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}agent_id'])!,
      visitId: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}visit_id']),
      lat: attachedDatabase.typeMapping
          .read(DriftSqlType.double, data['${effectivePrefix}lat'])!,
      lng: attachedDatabase.typeMapping
          .read(DriftSqlType.double, data['${effectivePrefix}lng'])!,
      accuracy: attachedDatabase.typeMapping
          .read(DriftSqlType.double, data['${effectivePrefix}accuracy']),
      recordedAt: attachedDatabase.typeMapping
          .read(DriftSqlType.dateTime, data['${effectivePrefix}recorded_at'])!,
      status: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}status'])!,
      error: attachedDatabase.typeMapping
          .read(DriftSqlType.string, data['${effectivePrefix}error']),
      createdAt: attachedDatabase.typeMapping
          .read(DriftSqlType.dateTime, data['${effectivePrefix}created_at'])!,
    );
  }

  @override
  $GpsPendingTable createAlias(String alias) {
    return $GpsPendingTable(attachedDatabase, alias);
  }
}

class GpsPendingRow extends DataClass implements Insertable<GpsPendingRow> {
  final String pointId;
  final String agentId;
  final String? visitId;
  final double lat;
  final double lng;
  final double? accuracy;
  final DateTime recordedAt;
  final String status;
  final String? error;
  final DateTime createdAt;
  const GpsPendingRow(
      {required this.pointId,
      required this.agentId,
      this.visitId,
      required this.lat,
      required this.lng,
      this.accuracy,
      required this.recordedAt,
      required this.status,
      this.error,
      required this.createdAt});
  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    map['point_id'] = Variable<String>(pointId);
    map['agent_id'] = Variable<String>(agentId);
    if (!nullToAbsent || visitId != null) {
      map['visit_id'] = Variable<String>(visitId);
    }
    map['lat'] = Variable<double>(lat);
    map['lng'] = Variable<double>(lng);
    if (!nullToAbsent || accuracy != null) {
      map['accuracy'] = Variable<double>(accuracy);
    }
    map['recorded_at'] = Variable<DateTime>(recordedAt);
    map['status'] = Variable<String>(status);
    if (!nullToAbsent || error != null) {
      map['error'] = Variable<String>(error);
    }
    map['created_at'] = Variable<DateTime>(createdAt);
    return map;
  }

  GpsPendingCompanion toCompanion(bool nullToAbsent) {
    return GpsPendingCompanion(
      pointId: Value(pointId),
      agentId: Value(agentId),
      visitId: visitId == null && nullToAbsent
          ? const Value.absent()
          : Value(visitId),
      lat: Value(lat),
      lng: Value(lng),
      accuracy: accuracy == null && nullToAbsent
          ? const Value.absent()
          : Value(accuracy),
      recordedAt: Value(recordedAt),
      status: Value(status),
      error:
          error == null && nullToAbsent ? const Value.absent() : Value(error),
      createdAt: Value(createdAt),
    );
  }

  factory GpsPendingRow.fromJson(Map<String, dynamic> json,
      {ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return GpsPendingRow(
      pointId: serializer.fromJson<String>(json['pointId']),
      agentId: serializer.fromJson<String>(json['agentId']),
      visitId: serializer.fromJson<String?>(json['visitId']),
      lat: serializer.fromJson<double>(json['lat']),
      lng: serializer.fromJson<double>(json['lng']),
      accuracy: serializer.fromJson<double?>(json['accuracy']),
      recordedAt: serializer.fromJson<DateTime>(json['recordedAt']),
      status: serializer.fromJson<String>(json['status']),
      error: serializer.fromJson<String?>(json['error']),
      createdAt: serializer.fromJson<DateTime>(json['createdAt']),
    );
  }
  @override
  Map<String, dynamic> toJson({ValueSerializer? serializer}) {
    serializer ??= driftRuntimeOptions.defaultSerializer;
    return <String, dynamic>{
      'pointId': serializer.toJson<String>(pointId),
      'agentId': serializer.toJson<String>(agentId),
      'visitId': serializer.toJson<String?>(visitId),
      'lat': serializer.toJson<double>(lat),
      'lng': serializer.toJson<double>(lng),
      'accuracy': serializer.toJson<double?>(accuracy),
      'recordedAt': serializer.toJson<DateTime>(recordedAt),
      'status': serializer.toJson<String>(status),
      'error': serializer.toJson<String?>(error),
      'createdAt': serializer.toJson<DateTime>(createdAt),
    };
  }

  GpsPendingRow copyWith(
          {String? pointId,
          String? agentId,
          Value<String?> visitId = const Value.absent(),
          double? lat,
          double? lng,
          Value<double?> accuracy = const Value.absent(),
          DateTime? recordedAt,
          String? status,
          Value<String?> error = const Value.absent(),
          DateTime? createdAt}) =>
      GpsPendingRow(
        pointId: pointId ?? this.pointId,
        agentId: agentId ?? this.agentId,
        visitId: visitId.present ? visitId.value : this.visitId,
        lat: lat ?? this.lat,
        lng: lng ?? this.lng,
        accuracy: accuracy.present ? accuracy.value : this.accuracy,
        recordedAt: recordedAt ?? this.recordedAt,
        status: status ?? this.status,
        error: error.present ? error.value : this.error,
        createdAt: createdAt ?? this.createdAt,
      );
  GpsPendingRow copyWithCompanion(GpsPendingCompanion data) {
    return GpsPendingRow(
      pointId: data.pointId.present ? data.pointId.value : this.pointId,
      agentId: data.agentId.present ? data.agentId.value : this.agentId,
      visitId: data.visitId.present ? data.visitId.value : this.visitId,
      lat: data.lat.present ? data.lat.value : this.lat,
      lng: data.lng.present ? data.lng.value : this.lng,
      accuracy: data.accuracy.present ? data.accuracy.value : this.accuracy,
      recordedAt:
          data.recordedAt.present ? data.recordedAt.value : this.recordedAt,
      status: data.status.present ? data.status.value : this.status,
      error: data.error.present ? data.error.value : this.error,
      createdAt: data.createdAt.present ? data.createdAt.value : this.createdAt,
    );
  }

  @override
  String toString() {
    return (StringBuffer('GpsPendingRow(')
          ..write('pointId: $pointId, ')
          ..write('agentId: $agentId, ')
          ..write('visitId: $visitId, ')
          ..write('lat: $lat, ')
          ..write('lng: $lng, ')
          ..write('accuracy: $accuracy, ')
          ..write('recordedAt: $recordedAt, ')
          ..write('status: $status, ')
          ..write('error: $error, ')
          ..write('createdAt: $createdAt')
          ..write(')'))
        .toString();
  }

  @override
  int get hashCode => Object.hash(pointId, agentId, visitId, lat, lng, accuracy,
      recordedAt, status, error, createdAt);
  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      (other is GpsPendingRow &&
          other.pointId == this.pointId &&
          other.agentId == this.agentId &&
          other.visitId == this.visitId &&
          other.lat == this.lat &&
          other.lng == this.lng &&
          other.accuracy == this.accuracy &&
          other.recordedAt == this.recordedAt &&
          other.status == this.status &&
          other.error == this.error &&
          other.createdAt == this.createdAt);
}

class GpsPendingCompanion extends UpdateCompanion<GpsPendingRow> {
  final Value<String> pointId;
  final Value<String> agentId;
  final Value<String?> visitId;
  final Value<double> lat;
  final Value<double> lng;
  final Value<double?> accuracy;
  final Value<DateTime> recordedAt;
  final Value<String> status;
  final Value<String?> error;
  final Value<DateTime> createdAt;
  final Value<int> rowid;
  const GpsPendingCompanion({
    this.pointId = const Value.absent(),
    this.agentId = const Value.absent(),
    this.visitId = const Value.absent(),
    this.lat = const Value.absent(),
    this.lng = const Value.absent(),
    this.accuracy = const Value.absent(),
    this.recordedAt = const Value.absent(),
    this.status = const Value.absent(),
    this.error = const Value.absent(),
    this.createdAt = const Value.absent(),
    this.rowid = const Value.absent(),
  });
  GpsPendingCompanion.insert({
    required String pointId,
    required String agentId,
    this.visitId = const Value.absent(),
    required double lat,
    required double lng,
    this.accuracy = const Value.absent(),
    required DateTime recordedAt,
    this.status = const Value.absent(),
    this.error = const Value.absent(),
    required DateTime createdAt,
    this.rowid = const Value.absent(),
  })  : pointId = Value(pointId),
        agentId = Value(agentId),
        lat = Value(lat),
        lng = Value(lng),
        recordedAt = Value(recordedAt),
        createdAt = Value(createdAt);
  static Insertable<GpsPendingRow> custom({
    Expression<String>? pointId,
    Expression<String>? agentId,
    Expression<String>? visitId,
    Expression<double>? lat,
    Expression<double>? lng,
    Expression<double>? accuracy,
    Expression<DateTime>? recordedAt,
    Expression<String>? status,
    Expression<String>? error,
    Expression<DateTime>? createdAt,
    Expression<int>? rowid,
  }) {
    return RawValuesInsertable({
      if (pointId != null) 'point_id': pointId,
      if (agentId != null) 'agent_id': agentId,
      if (visitId != null) 'visit_id': visitId,
      if (lat != null) 'lat': lat,
      if (lng != null) 'lng': lng,
      if (accuracy != null) 'accuracy': accuracy,
      if (recordedAt != null) 'recorded_at': recordedAt,
      if (status != null) 'status': status,
      if (error != null) 'error': error,
      if (createdAt != null) 'created_at': createdAt,
      if (rowid != null) 'rowid': rowid,
    });
  }

  GpsPendingCompanion copyWith(
      {Value<String>? pointId,
      Value<String>? agentId,
      Value<String?>? visitId,
      Value<double>? lat,
      Value<double>? lng,
      Value<double?>? accuracy,
      Value<DateTime>? recordedAt,
      Value<String>? status,
      Value<String?>? error,
      Value<DateTime>? createdAt,
      Value<int>? rowid}) {
    return GpsPendingCompanion(
      pointId: pointId ?? this.pointId,
      agentId: agentId ?? this.agentId,
      visitId: visitId ?? this.visitId,
      lat: lat ?? this.lat,
      lng: lng ?? this.lng,
      accuracy: accuracy ?? this.accuracy,
      recordedAt: recordedAt ?? this.recordedAt,
      status: status ?? this.status,
      error: error ?? this.error,
      createdAt: createdAt ?? this.createdAt,
      rowid: rowid ?? this.rowid,
    );
  }

  @override
  Map<String, Expression> toColumns(bool nullToAbsent) {
    final map = <String, Expression>{};
    if (pointId.present) {
      map['point_id'] = Variable<String>(pointId.value);
    }
    if (agentId.present) {
      map['agent_id'] = Variable<String>(agentId.value);
    }
    if (visitId.present) {
      map['visit_id'] = Variable<String>(visitId.value);
    }
    if (lat.present) {
      map['lat'] = Variable<double>(lat.value);
    }
    if (lng.present) {
      map['lng'] = Variable<double>(lng.value);
    }
    if (accuracy.present) {
      map['accuracy'] = Variable<double>(accuracy.value);
    }
    if (recordedAt.present) {
      map['recorded_at'] = Variable<DateTime>(recordedAt.value);
    }
    if (status.present) {
      map['status'] = Variable<String>(status.value);
    }
    if (error.present) {
      map['error'] = Variable<String>(error.value);
    }
    if (createdAt.present) {
      map['created_at'] = Variable<DateTime>(createdAt.value);
    }
    if (rowid.present) {
      map['rowid'] = Variable<int>(rowid.value);
    }
    return map;
  }

  @override
  String toString() {
    return (StringBuffer('GpsPendingCompanion(')
          ..write('pointId: $pointId, ')
          ..write('agentId: $agentId, ')
          ..write('visitId: $visitId, ')
          ..write('lat: $lat, ')
          ..write('lng: $lng, ')
          ..write('accuracy: $accuracy, ')
          ..write('recordedAt: $recordedAt, ')
          ..write('status: $status, ')
          ..write('error: $error, ')
          ..write('createdAt: $createdAt, ')
          ..write('rowid: $rowid')
          ..write(')'))
        .toString();
  }
}

abstract class _$SfaDatabase extends GeneratedDatabase {
  _$SfaDatabase(QueryExecutor e) : super(e);
  $SfaDatabaseManager get managers => $SfaDatabaseManager(this);
  late final $CachedEntitiesTable cachedEntities = $CachedEntitiesTable(this);
  late final $SyncMetaTable syncMeta = $SyncMetaTable(this);
  late final $OutboxOpsTable outboxOps = $OutboxOpsTable(this);
  late final $FileUploadsTable fileUploads = $FileUploadsTable(this);
  late final $GpsPendingTable gpsPending = $GpsPendingTable(this);
  @override
  Iterable<TableInfo<Table, Object?>> get allTables =>
      allSchemaEntities.whereType<TableInfo<Table, Object?>>();
  @override
  List<DatabaseSchemaEntity> get allSchemaEntities =>
      [cachedEntities, syncMeta, outboxOps, fileUploads, gpsPending];
}

typedef $$CachedEntitiesTableCreateCompanionBuilder = CachedEntitiesCompanion
    Function({
  required String entityType,
  required String entityId,
  required String payloadJson,
  required int updatedAt,
  Value<int> rowid,
});
typedef $$CachedEntitiesTableUpdateCompanionBuilder = CachedEntitiesCompanion
    Function({
  Value<String> entityType,
  Value<String> entityId,
  Value<String> payloadJson,
  Value<int> updatedAt,
  Value<int> rowid,
});

class $$CachedEntitiesTableFilterComposer
    extends Composer<_$SfaDatabase, $CachedEntitiesTable> {
  $$CachedEntitiesTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get entityType => $composableBuilder(
      column: $table.entityType, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get entityId => $composableBuilder(
      column: $table.entityId, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get payloadJson => $composableBuilder(
      column: $table.payloadJson, builder: (column) => ColumnFilters(column));

  ColumnFilters<int> get updatedAt => $composableBuilder(
      column: $table.updatedAt, builder: (column) => ColumnFilters(column));
}

class $$CachedEntitiesTableOrderingComposer
    extends Composer<_$SfaDatabase, $CachedEntitiesTable> {
  $$CachedEntitiesTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get entityType => $composableBuilder(
      column: $table.entityType, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get entityId => $composableBuilder(
      column: $table.entityId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get payloadJson => $composableBuilder(
      column: $table.payloadJson, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<int> get updatedAt => $composableBuilder(
      column: $table.updatedAt, builder: (column) => ColumnOrderings(column));
}

class $$CachedEntitiesTableAnnotationComposer
    extends Composer<_$SfaDatabase, $CachedEntitiesTable> {
  $$CachedEntitiesTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get entityType => $composableBuilder(
      column: $table.entityType, builder: (column) => column);

  GeneratedColumn<String> get entityId =>
      $composableBuilder(column: $table.entityId, builder: (column) => column);

  GeneratedColumn<String> get payloadJson => $composableBuilder(
      column: $table.payloadJson, builder: (column) => column);

  GeneratedColumn<int> get updatedAt =>
      $composableBuilder(column: $table.updatedAt, builder: (column) => column);
}

class $$CachedEntitiesTableTableManager extends RootTableManager<
    _$SfaDatabase,
    $CachedEntitiesTable,
    CachedEntityRow,
    $$CachedEntitiesTableFilterComposer,
    $$CachedEntitiesTableOrderingComposer,
    $$CachedEntitiesTableAnnotationComposer,
    $$CachedEntitiesTableCreateCompanionBuilder,
    $$CachedEntitiesTableUpdateCompanionBuilder,
    (
      CachedEntityRow,
      BaseReferences<_$SfaDatabase, $CachedEntitiesTable, CachedEntityRow>
    ),
    CachedEntityRow,
    PrefetchHooks Function()> {
  $$CachedEntitiesTableTableManager(
      _$SfaDatabase db, $CachedEntitiesTable table)
      : super(TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$CachedEntitiesTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$CachedEntitiesTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$CachedEntitiesTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback: ({
            Value<String> entityType = const Value.absent(),
            Value<String> entityId = const Value.absent(),
            Value<String> payloadJson = const Value.absent(),
            Value<int> updatedAt = const Value.absent(),
            Value<int> rowid = const Value.absent(),
          }) =>
              CachedEntitiesCompanion(
            entityType: entityType,
            entityId: entityId,
            payloadJson: payloadJson,
            updatedAt: updatedAt,
            rowid: rowid,
          ),
          createCompanionCallback: ({
            required String entityType,
            required String entityId,
            required String payloadJson,
            required int updatedAt,
            Value<int> rowid = const Value.absent(),
          }) =>
              CachedEntitiesCompanion.insert(
            entityType: entityType,
            entityId: entityId,
            payloadJson: payloadJson,
            updatedAt: updatedAt,
            rowid: rowid,
          ),
          withReferenceMapper: (p0) => p0
              .map((e) => (e.readTable(table), BaseReferences(db, table, e)))
              .toList(),
          prefetchHooksCallback: null,
        ));
}

typedef $$CachedEntitiesTableProcessedTableManager = ProcessedTableManager<
    _$SfaDatabase,
    $CachedEntitiesTable,
    CachedEntityRow,
    $$CachedEntitiesTableFilterComposer,
    $$CachedEntitiesTableOrderingComposer,
    $$CachedEntitiesTableAnnotationComposer,
    $$CachedEntitiesTableCreateCompanionBuilder,
    $$CachedEntitiesTableUpdateCompanionBuilder,
    (
      CachedEntityRow,
      BaseReferences<_$SfaDatabase, $CachedEntitiesTable, CachedEntityRow>
    ),
    CachedEntityRow,
    PrefetchHooks Function()>;
typedef $$SyncMetaTableCreateCompanionBuilder = SyncMetaCompanion Function({
  required String key,
  required String value,
  Value<int> rowid,
});
typedef $$SyncMetaTableUpdateCompanionBuilder = SyncMetaCompanion Function({
  Value<String> key,
  Value<String> value,
  Value<int> rowid,
});

class $$SyncMetaTableFilterComposer
    extends Composer<_$SfaDatabase, $SyncMetaTable> {
  $$SyncMetaTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get key => $composableBuilder(
      column: $table.key, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get value => $composableBuilder(
      column: $table.value, builder: (column) => ColumnFilters(column));
}

class $$SyncMetaTableOrderingComposer
    extends Composer<_$SfaDatabase, $SyncMetaTable> {
  $$SyncMetaTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get key => $composableBuilder(
      column: $table.key, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get value => $composableBuilder(
      column: $table.value, builder: (column) => ColumnOrderings(column));
}

class $$SyncMetaTableAnnotationComposer
    extends Composer<_$SfaDatabase, $SyncMetaTable> {
  $$SyncMetaTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get key =>
      $composableBuilder(column: $table.key, builder: (column) => column);

  GeneratedColumn<String> get value =>
      $composableBuilder(column: $table.value, builder: (column) => column);
}

class $$SyncMetaTableTableManager extends RootTableManager<
    _$SfaDatabase,
    $SyncMetaTable,
    SyncMetaData,
    $$SyncMetaTableFilterComposer,
    $$SyncMetaTableOrderingComposer,
    $$SyncMetaTableAnnotationComposer,
    $$SyncMetaTableCreateCompanionBuilder,
    $$SyncMetaTableUpdateCompanionBuilder,
    (SyncMetaData, BaseReferences<_$SfaDatabase, $SyncMetaTable, SyncMetaData>),
    SyncMetaData,
    PrefetchHooks Function()> {
  $$SyncMetaTableTableManager(_$SfaDatabase db, $SyncMetaTable table)
      : super(TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$SyncMetaTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$SyncMetaTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$SyncMetaTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback: ({
            Value<String> key = const Value.absent(),
            Value<String> value = const Value.absent(),
            Value<int> rowid = const Value.absent(),
          }) =>
              SyncMetaCompanion(
            key: key,
            value: value,
            rowid: rowid,
          ),
          createCompanionCallback: ({
            required String key,
            required String value,
            Value<int> rowid = const Value.absent(),
          }) =>
              SyncMetaCompanion.insert(
            key: key,
            value: value,
            rowid: rowid,
          ),
          withReferenceMapper: (p0) => p0
              .map((e) => (e.readTable(table), BaseReferences(db, table, e)))
              .toList(),
          prefetchHooksCallback: null,
        ));
}

typedef $$SyncMetaTableProcessedTableManager = ProcessedTableManager<
    _$SfaDatabase,
    $SyncMetaTable,
    SyncMetaData,
    $$SyncMetaTableFilterComposer,
    $$SyncMetaTableOrderingComposer,
    $$SyncMetaTableAnnotationComposer,
    $$SyncMetaTableCreateCompanionBuilder,
    $$SyncMetaTableUpdateCompanionBuilder,
    (SyncMetaData, BaseReferences<_$SfaDatabase, $SyncMetaTable, SyncMetaData>),
    SyncMetaData,
    PrefetchHooks Function()>;
typedef $$OutboxOpsTableCreateCompanionBuilder = OutboxOpsCompanion Function({
  required String opId,
  required String entityType,
  required String entityId,
  required String op,
  Value<int> baseVersion,
  required String payloadJson,
  Value<String> status,
  required DateTime createdAt,
  Value<int> rowid,
});
typedef $$OutboxOpsTableUpdateCompanionBuilder = OutboxOpsCompanion Function({
  Value<String> opId,
  Value<String> entityType,
  Value<String> entityId,
  Value<String> op,
  Value<int> baseVersion,
  Value<String> payloadJson,
  Value<String> status,
  Value<DateTime> createdAt,
  Value<int> rowid,
});

class $$OutboxOpsTableFilterComposer
    extends Composer<_$SfaDatabase, $OutboxOpsTable> {
  $$OutboxOpsTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get opId => $composableBuilder(
      column: $table.opId, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get entityType => $composableBuilder(
      column: $table.entityType, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get entityId => $composableBuilder(
      column: $table.entityId, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get op => $composableBuilder(
      column: $table.op, builder: (column) => ColumnFilters(column));

  ColumnFilters<int> get baseVersion => $composableBuilder(
      column: $table.baseVersion, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get payloadJson => $composableBuilder(
      column: $table.payloadJson, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get status => $composableBuilder(
      column: $table.status, builder: (column) => ColumnFilters(column));

  ColumnFilters<DateTime> get createdAt => $composableBuilder(
      column: $table.createdAt, builder: (column) => ColumnFilters(column));
}

class $$OutboxOpsTableOrderingComposer
    extends Composer<_$SfaDatabase, $OutboxOpsTable> {
  $$OutboxOpsTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get opId => $composableBuilder(
      column: $table.opId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get entityType => $composableBuilder(
      column: $table.entityType, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get entityId => $composableBuilder(
      column: $table.entityId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get op => $composableBuilder(
      column: $table.op, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<int> get baseVersion => $composableBuilder(
      column: $table.baseVersion, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get payloadJson => $composableBuilder(
      column: $table.payloadJson, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get status => $composableBuilder(
      column: $table.status, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<DateTime> get createdAt => $composableBuilder(
      column: $table.createdAt, builder: (column) => ColumnOrderings(column));
}

class $$OutboxOpsTableAnnotationComposer
    extends Composer<_$SfaDatabase, $OutboxOpsTable> {
  $$OutboxOpsTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get opId =>
      $composableBuilder(column: $table.opId, builder: (column) => column);

  GeneratedColumn<String> get entityType => $composableBuilder(
      column: $table.entityType, builder: (column) => column);

  GeneratedColumn<String> get entityId =>
      $composableBuilder(column: $table.entityId, builder: (column) => column);

  GeneratedColumn<String> get op =>
      $composableBuilder(column: $table.op, builder: (column) => column);

  GeneratedColumn<int> get baseVersion => $composableBuilder(
      column: $table.baseVersion, builder: (column) => column);

  GeneratedColumn<String> get payloadJson => $composableBuilder(
      column: $table.payloadJson, builder: (column) => column);

  GeneratedColumn<String> get status =>
      $composableBuilder(column: $table.status, builder: (column) => column);

  GeneratedColumn<DateTime> get createdAt =>
      $composableBuilder(column: $table.createdAt, builder: (column) => column);
}

class $$OutboxOpsTableTableManager extends RootTableManager<
    _$SfaDatabase,
    $OutboxOpsTable,
    OutboxOpRow,
    $$OutboxOpsTableFilterComposer,
    $$OutboxOpsTableOrderingComposer,
    $$OutboxOpsTableAnnotationComposer,
    $$OutboxOpsTableCreateCompanionBuilder,
    $$OutboxOpsTableUpdateCompanionBuilder,
    (OutboxOpRow, BaseReferences<_$SfaDatabase, $OutboxOpsTable, OutboxOpRow>),
    OutboxOpRow,
    PrefetchHooks Function()> {
  $$OutboxOpsTableTableManager(_$SfaDatabase db, $OutboxOpsTable table)
      : super(TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$OutboxOpsTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$OutboxOpsTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$OutboxOpsTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback: ({
            Value<String> opId = const Value.absent(),
            Value<String> entityType = const Value.absent(),
            Value<String> entityId = const Value.absent(),
            Value<String> op = const Value.absent(),
            Value<int> baseVersion = const Value.absent(),
            Value<String> payloadJson = const Value.absent(),
            Value<String> status = const Value.absent(),
            Value<DateTime> createdAt = const Value.absent(),
            Value<int> rowid = const Value.absent(),
          }) =>
              OutboxOpsCompanion(
            opId: opId,
            entityType: entityType,
            entityId: entityId,
            op: op,
            baseVersion: baseVersion,
            payloadJson: payloadJson,
            status: status,
            createdAt: createdAt,
            rowid: rowid,
          ),
          createCompanionCallback: ({
            required String opId,
            required String entityType,
            required String entityId,
            required String op,
            Value<int> baseVersion = const Value.absent(),
            required String payloadJson,
            Value<String> status = const Value.absent(),
            required DateTime createdAt,
            Value<int> rowid = const Value.absent(),
          }) =>
              OutboxOpsCompanion.insert(
            opId: opId,
            entityType: entityType,
            entityId: entityId,
            op: op,
            baseVersion: baseVersion,
            payloadJson: payloadJson,
            status: status,
            createdAt: createdAt,
            rowid: rowid,
          ),
          withReferenceMapper: (p0) => p0
              .map((e) => (e.readTable(table), BaseReferences(db, table, e)))
              .toList(),
          prefetchHooksCallback: null,
        ));
}

typedef $$OutboxOpsTableProcessedTableManager = ProcessedTableManager<
    _$SfaDatabase,
    $OutboxOpsTable,
    OutboxOpRow,
    $$OutboxOpsTableFilterComposer,
    $$OutboxOpsTableOrderingComposer,
    $$OutboxOpsTableAnnotationComposer,
    $$OutboxOpsTableCreateCompanionBuilder,
    $$OutboxOpsTableUpdateCompanionBuilder,
    (OutboxOpRow, BaseReferences<_$SfaDatabase, $OutboxOpsTable, OutboxOpRow>),
    OutboxOpRow,
    PrefetchHooks Function()>;
typedef $$FileUploadsTableCreateCompanionBuilder = FileUploadsCompanion
    Function({
  required String uploadId,
  required String fileName,
  required String mime,
  Value<int> sizeBytes,
  Value<String?> localPath,
  Value<String> status,
  Value<String?> remoteFileId,
  Value<String?> error,
  required DateTime createdAt,
  Value<Uint8List?> payload,
  Value<int> rowid,
});
typedef $$FileUploadsTableUpdateCompanionBuilder = FileUploadsCompanion
    Function({
  Value<String> uploadId,
  Value<String> fileName,
  Value<String> mime,
  Value<int> sizeBytes,
  Value<String?> localPath,
  Value<String> status,
  Value<String?> remoteFileId,
  Value<String?> error,
  Value<DateTime> createdAt,
  Value<Uint8List?> payload,
  Value<int> rowid,
});

class $$FileUploadsTableFilterComposer
    extends Composer<_$SfaDatabase, $FileUploadsTable> {
  $$FileUploadsTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get uploadId => $composableBuilder(
      column: $table.uploadId, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get fileName => $composableBuilder(
      column: $table.fileName, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get mime => $composableBuilder(
      column: $table.mime, builder: (column) => ColumnFilters(column));

  ColumnFilters<int> get sizeBytes => $composableBuilder(
      column: $table.sizeBytes, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get localPath => $composableBuilder(
      column: $table.localPath, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get status => $composableBuilder(
      column: $table.status, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get remoteFileId => $composableBuilder(
      column: $table.remoteFileId, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get error => $composableBuilder(
      column: $table.error, builder: (column) => ColumnFilters(column));

  ColumnFilters<DateTime> get createdAt => $composableBuilder(
      column: $table.createdAt, builder: (column) => ColumnFilters(column));

  ColumnFilters<Uint8List> get payload => $composableBuilder(
      column: $table.payload, builder: (column) => ColumnFilters(column));
}

class $$FileUploadsTableOrderingComposer
    extends Composer<_$SfaDatabase, $FileUploadsTable> {
  $$FileUploadsTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get uploadId => $composableBuilder(
      column: $table.uploadId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get fileName => $composableBuilder(
      column: $table.fileName, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get mime => $composableBuilder(
      column: $table.mime, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<int> get sizeBytes => $composableBuilder(
      column: $table.sizeBytes, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get localPath => $composableBuilder(
      column: $table.localPath, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get status => $composableBuilder(
      column: $table.status, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get remoteFileId => $composableBuilder(
      column: $table.remoteFileId,
      builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get error => $composableBuilder(
      column: $table.error, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<DateTime> get createdAt => $composableBuilder(
      column: $table.createdAt, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<Uint8List> get payload => $composableBuilder(
      column: $table.payload, builder: (column) => ColumnOrderings(column));
}

class $$FileUploadsTableAnnotationComposer
    extends Composer<_$SfaDatabase, $FileUploadsTable> {
  $$FileUploadsTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get uploadId =>
      $composableBuilder(column: $table.uploadId, builder: (column) => column);

  GeneratedColumn<String> get fileName =>
      $composableBuilder(column: $table.fileName, builder: (column) => column);

  GeneratedColumn<String> get mime =>
      $composableBuilder(column: $table.mime, builder: (column) => column);

  GeneratedColumn<int> get sizeBytes =>
      $composableBuilder(column: $table.sizeBytes, builder: (column) => column);

  GeneratedColumn<String> get localPath =>
      $composableBuilder(column: $table.localPath, builder: (column) => column);

  GeneratedColumn<String> get status =>
      $composableBuilder(column: $table.status, builder: (column) => column);

  GeneratedColumn<String> get remoteFileId => $composableBuilder(
      column: $table.remoteFileId, builder: (column) => column);

  GeneratedColumn<String> get error =>
      $composableBuilder(column: $table.error, builder: (column) => column);

  GeneratedColumn<DateTime> get createdAt =>
      $composableBuilder(column: $table.createdAt, builder: (column) => column);

  GeneratedColumn<Uint8List> get payload =>
      $composableBuilder(column: $table.payload, builder: (column) => column);
}

class $$FileUploadsTableTableManager extends RootTableManager<
    _$SfaDatabase,
    $FileUploadsTable,
    FileUploadRow,
    $$FileUploadsTableFilterComposer,
    $$FileUploadsTableOrderingComposer,
    $$FileUploadsTableAnnotationComposer,
    $$FileUploadsTableCreateCompanionBuilder,
    $$FileUploadsTableUpdateCompanionBuilder,
    (
      FileUploadRow,
      BaseReferences<_$SfaDatabase, $FileUploadsTable, FileUploadRow>
    ),
    FileUploadRow,
    PrefetchHooks Function()> {
  $$FileUploadsTableTableManager(_$SfaDatabase db, $FileUploadsTable table)
      : super(TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$FileUploadsTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$FileUploadsTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$FileUploadsTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback: ({
            Value<String> uploadId = const Value.absent(),
            Value<String> fileName = const Value.absent(),
            Value<String> mime = const Value.absent(),
            Value<int> sizeBytes = const Value.absent(),
            Value<String?> localPath = const Value.absent(),
            Value<String> status = const Value.absent(),
            Value<String?> remoteFileId = const Value.absent(),
            Value<String?> error = const Value.absent(),
            Value<DateTime> createdAt = const Value.absent(),
            Value<Uint8List?> payload = const Value.absent(),
            Value<int> rowid = const Value.absent(),
          }) =>
              FileUploadsCompanion(
            uploadId: uploadId,
            fileName: fileName,
            mime: mime,
            sizeBytes: sizeBytes,
            localPath: localPath,
            status: status,
            remoteFileId: remoteFileId,
            error: error,
            createdAt: createdAt,
            payload: payload,
            rowid: rowid,
          ),
          createCompanionCallback: ({
            required String uploadId,
            required String fileName,
            required String mime,
            Value<int> sizeBytes = const Value.absent(),
            Value<String?> localPath = const Value.absent(),
            Value<String> status = const Value.absent(),
            Value<String?> remoteFileId = const Value.absent(),
            Value<String?> error = const Value.absent(),
            required DateTime createdAt,
            Value<Uint8List?> payload = const Value.absent(),
            Value<int> rowid = const Value.absent(),
          }) =>
              FileUploadsCompanion.insert(
            uploadId: uploadId,
            fileName: fileName,
            mime: mime,
            sizeBytes: sizeBytes,
            localPath: localPath,
            status: status,
            remoteFileId: remoteFileId,
            error: error,
            createdAt: createdAt,
            payload: payload,
            rowid: rowid,
          ),
          withReferenceMapper: (p0) => p0
              .map((e) => (e.readTable(table), BaseReferences(db, table, e)))
              .toList(),
          prefetchHooksCallback: null,
        ));
}

typedef $$FileUploadsTableProcessedTableManager = ProcessedTableManager<
    _$SfaDatabase,
    $FileUploadsTable,
    FileUploadRow,
    $$FileUploadsTableFilterComposer,
    $$FileUploadsTableOrderingComposer,
    $$FileUploadsTableAnnotationComposer,
    $$FileUploadsTableCreateCompanionBuilder,
    $$FileUploadsTableUpdateCompanionBuilder,
    (
      FileUploadRow,
      BaseReferences<_$SfaDatabase, $FileUploadsTable, FileUploadRow>
    ),
    FileUploadRow,
    PrefetchHooks Function()>;
typedef $$GpsPendingTableCreateCompanionBuilder = GpsPendingCompanion Function({
  required String pointId,
  required String agentId,
  Value<String?> visitId,
  required double lat,
  required double lng,
  Value<double?> accuracy,
  required DateTime recordedAt,
  Value<String> status,
  Value<String?> error,
  required DateTime createdAt,
  Value<int> rowid,
});
typedef $$GpsPendingTableUpdateCompanionBuilder = GpsPendingCompanion Function({
  Value<String> pointId,
  Value<String> agentId,
  Value<String?> visitId,
  Value<double> lat,
  Value<double> lng,
  Value<double?> accuracy,
  Value<DateTime> recordedAt,
  Value<String> status,
  Value<String?> error,
  Value<DateTime> createdAt,
  Value<int> rowid,
});

class $$GpsPendingTableFilterComposer
    extends Composer<_$SfaDatabase, $GpsPendingTable> {
  $$GpsPendingTableFilterComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnFilters<String> get pointId => $composableBuilder(
      column: $table.pointId, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get agentId => $composableBuilder(
      column: $table.agentId, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get visitId => $composableBuilder(
      column: $table.visitId, builder: (column) => ColumnFilters(column));

  ColumnFilters<double> get lat => $composableBuilder(
      column: $table.lat, builder: (column) => ColumnFilters(column));

  ColumnFilters<double> get lng => $composableBuilder(
      column: $table.lng, builder: (column) => ColumnFilters(column));

  ColumnFilters<double> get accuracy => $composableBuilder(
      column: $table.accuracy, builder: (column) => ColumnFilters(column));

  ColumnFilters<DateTime> get recordedAt => $composableBuilder(
      column: $table.recordedAt, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get status => $composableBuilder(
      column: $table.status, builder: (column) => ColumnFilters(column));

  ColumnFilters<String> get error => $composableBuilder(
      column: $table.error, builder: (column) => ColumnFilters(column));

  ColumnFilters<DateTime> get createdAt => $composableBuilder(
      column: $table.createdAt, builder: (column) => ColumnFilters(column));
}

class $$GpsPendingTableOrderingComposer
    extends Composer<_$SfaDatabase, $GpsPendingTable> {
  $$GpsPendingTableOrderingComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  ColumnOrderings<String> get pointId => $composableBuilder(
      column: $table.pointId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get agentId => $composableBuilder(
      column: $table.agentId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get visitId => $composableBuilder(
      column: $table.visitId, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<double> get lat => $composableBuilder(
      column: $table.lat, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<double> get lng => $composableBuilder(
      column: $table.lng, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<double> get accuracy => $composableBuilder(
      column: $table.accuracy, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<DateTime> get recordedAt => $composableBuilder(
      column: $table.recordedAt, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get status => $composableBuilder(
      column: $table.status, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<String> get error => $composableBuilder(
      column: $table.error, builder: (column) => ColumnOrderings(column));

  ColumnOrderings<DateTime> get createdAt => $composableBuilder(
      column: $table.createdAt, builder: (column) => ColumnOrderings(column));
}

class $$GpsPendingTableAnnotationComposer
    extends Composer<_$SfaDatabase, $GpsPendingTable> {
  $$GpsPendingTableAnnotationComposer({
    required super.$db,
    required super.$table,
    super.joinBuilder,
    super.$addJoinBuilderToRootComposer,
    super.$removeJoinBuilderFromRootComposer,
  });
  GeneratedColumn<String> get pointId =>
      $composableBuilder(column: $table.pointId, builder: (column) => column);

  GeneratedColumn<String> get agentId =>
      $composableBuilder(column: $table.agentId, builder: (column) => column);

  GeneratedColumn<String> get visitId =>
      $composableBuilder(column: $table.visitId, builder: (column) => column);

  GeneratedColumn<double> get lat =>
      $composableBuilder(column: $table.lat, builder: (column) => column);

  GeneratedColumn<double> get lng =>
      $composableBuilder(column: $table.lng, builder: (column) => column);

  GeneratedColumn<double> get accuracy =>
      $composableBuilder(column: $table.accuracy, builder: (column) => column);

  GeneratedColumn<DateTime> get recordedAt => $composableBuilder(
      column: $table.recordedAt, builder: (column) => column);

  GeneratedColumn<String> get status =>
      $composableBuilder(column: $table.status, builder: (column) => column);

  GeneratedColumn<String> get error =>
      $composableBuilder(column: $table.error, builder: (column) => column);

  GeneratedColumn<DateTime> get createdAt =>
      $composableBuilder(column: $table.createdAt, builder: (column) => column);
}

class $$GpsPendingTableTableManager extends RootTableManager<
    _$SfaDatabase,
    $GpsPendingTable,
    GpsPendingRow,
    $$GpsPendingTableFilterComposer,
    $$GpsPendingTableOrderingComposer,
    $$GpsPendingTableAnnotationComposer,
    $$GpsPendingTableCreateCompanionBuilder,
    $$GpsPendingTableUpdateCompanionBuilder,
    (
      GpsPendingRow,
      BaseReferences<_$SfaDatabase, $GpsPendingTable, GpsPendingRow>
    ),
    GpsPendingRow,
    PrefetchHooks Function()> {
  $$GpsPendingTableTableManager(_$SfaDatabase db, $GpsPendingTable table)
      : super(TableManagerState(
          db: db,
          table: table,
          createFilteringComposer: () =>
              $$GpsPendingTableFilterComposer($db: db, $table: table),
          createOrderingComposer: () =>
              $$GpsPendingTableOrderingComposer($db: db, $table: table),
          createComputedFieldComposer: () =>
              $$GpsPendingTableAnnotationComposer($db: db, $table: table),
          updateCompanionCallback: ({
            Value<String> pointId = const Value.absent(),
            Value<String> agentId = const Value.absent(),
            Value<String?> visitId = const Value.absent(),
            Value<double> lat = const Value.absent(),
            Value<double> lng = const Value.absent(),
            Value<double?> accuracy = const Value.absent(),
            Value<DateTime> recordedAt = const Value.absent(),
            Value<String> status = const Value.absent(),
            Value<String?> error = const Value.absent(),
            Value<DateTime> createdAt = const Value.absent(),
            Value<int> rowid = const Value.absent(),
          }) =>
              GpsPendingCompanion(
            pointId: pointId,
            agentId: agentId,
            visitId: visitId,
            lat: lat,
            lng: lng,
            accuracy: accuracy,
            recordedAt: recordedAt,
            status: status,
            error: error,
            createdAt: createdAt,
            rowid: rowid,
          ),
          createCompanionCallback: ({
            required String pointId,
            required String agentId,
            Value<String?> visitId = const Value.absent(),
            required double lat,
            required double lng,
            Value<double?> accuracy = const Value.absent(),
            required DateTime recordedAt,
            Value<String> status = const Value.absent(),
            Value<String?> error = const Value.absent(),
            required DateTime createdAt,
            Value<int> rowid = const Value.absent(),
          }) =>
              GpsPendingCompanion.insert(
            pointId: pointId,
            agentId: agentId,
            visitId: visitId,
            lat: lat,
            lng: lng,
            accuracy: accuracy,
            recordedAt: recordedAt,
            status: status,
            error: error,
            createdAt: createdAt,
            rowid: rowid,
          ),
          withReferenceMapper: (p0) => p0
              .map((e) => (e.readTable(table), BaseReferences(db, table, e)))
              .toList(),
          prefetchHooksCallback: null,
        ));
}

typedef $$GpsPendingTableProcessedTableManager = ProcessedTableManager<
    _$SfaDatabase,
    $GpsPendingTable,
    GpsPendingRow,
    $$GpsPendingTableFilterComposer,
    $$GpsPendingTableOrderingComposer,
    $$GpsPendingTableAnnotationComposer,
    $$GpsPendingTableCreateCompanionBuilder,
    $$GpsPendingTableUpdateCompanionBuilder,
    (
      GpsPendingRow,
      BaseReferences<_$SfaDatabase, $GpsPendingTable, GpsPendingRow>
    ),
    GpsPendingRow,
    PrefetchHooks Function()>;

class $SfaDatabaseManager {
  final _$SfaDatabase _db;
  $SfaDatabaseManager(this._db);
  $$CachedEntitiesTableTableManager get cachedEntities =>
      $$CachedEntitiesTableTableManager(_db, _db.cachedEntities);
  $$SyncMetaTableTableManager get syncMeta =>
      $$SyncMetaTableTableManager(_db, _db.syncMeta);
  $$OutboxOpsTableTableManager get outboxOps =>
      $$OutboxOpsTableTableManager(_db, _db.outboxOps);
  $$FileUploadsTableTableManager get fileUploads =>
      $$FileUploadsTableTableManager(_db, _db.fileUploads);
  $$GpsPendingTableTableManager get gpsPending =>
      $$GpsPendingTableTableManager(_db, _db.gpsPending);
}
