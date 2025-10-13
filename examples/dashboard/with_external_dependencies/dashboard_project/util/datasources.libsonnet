local g = import './g.libsonnet';
local envs = import './envs.libsonnet';

local datasource =  g.dashboard.annotation.datasource;

{
    base(type, uid):
        datasource.withType(type)
        + datasource.withUid(uid),

    dashboardDatasource:
        self.base('datasource', 'grafana'),

    prometheusDatasource:
        self.base(envs.PROMETHEUS_DS_TYPE, envs.PROMETHEUS_DS_UID),
}
