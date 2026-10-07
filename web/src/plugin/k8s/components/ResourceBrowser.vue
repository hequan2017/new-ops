<template>
  <el-drawer v-model="visible" :title="`资源浏览 · ${clusterName}`" size="75%" destroy-on-close>
    <div class="ops-btn-list" style="margin-bottom: 8px">
      <span style="font-size: 13px; color: #606266">命名空间：</span>
      <el-select v-model="nsFilter" size="small" style="width: 220px" clearable placeholder="全部（授权范围）" @change="loadAll">
        <el-option v-for="n in nsOptions" :key="n.name" :label="n.name" :value="n.name" />
      </el-select>
    </div>
    <el-tabs v-model="activeTab">
      <el-tab-pane label="Pods" name="pods">
        <el-table :data="pods" size="small" v-loading="loading">
          <el-table-column prop="name" label="名称" min-width="200" />
          <el-table-column prop="namespace" label="命名空间" min-width="110" />
          <el-table-column prop="status" label="状态" width="110">
            <template #default="{ row }">
              <el-tag :type="row.status === 'Running' ? 'success' : 'warning'" size="small">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="restarts" label="重启" width="70" />
          <el-table-column prop="node" label="节点" min-width="130" />
          <el-table-column prop="age" label="年龄" width="80" />
          <el-table-column label="操作" width="260" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" icon="document" @click="showLogs(row)">日志</el-button>
              <el-button link type="primary" icon="view" @click="showPodDetail(row)">详情</el-button>
              <el-button link type="success" icon="monitor" @click="openShell(row)">终端</el-button>
              <el-button link type="danger" icon="delete" @click="onDeletePod(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
      <el-tab-pane label="Deployments" name="deployments">
        <el-table :data="deployments" size="small" v-loading="loading">
          <el-table-column prop="name" label="名称" min-width="160" />
          <el-table-column prop="namespace" label="命名空间" min-width="110" />
          <el-table-column label="副本" width="120">
            <template #default="{ row }">{{ row.ready }}/{{ row.replicas }}</template>
          </el-table-column>
          <el-table-column prop="age" label="年龄" width="90" />
          <el-table-column label="操作" width="230" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" icon="sort" @click="$emit('scale', row)">扩缩容</el-button>
              <el-button link type="warning" icon="refresh" @click="$emit('restart', row)">重启</el-button>
              <el-button link type="info" icon="tickets" @click="showYaml('deployment', row)">YAML</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
      <el-tab-pane label="StatefulSets" name="statefulsets">
        <el-table :data="statefulsets" size="small" v-loading="loading">
          <el-table-column prop="name" label="名称" min-width="160" />
          <el-table-column prop="namespace" label="命名空间" min-width="110" />
          <el-table-column label="就绪" width="100">
            <template #default="{ row }">{{ row.ready }}/{{ row.replicas }}</template>
          </el-table-column>
          <el-table-column prop="age" label="年龄" width="90" />
          <el-table-column label="操作" width="90" fixed="right">
            <template #default="{ row }">
              <el-button link type="info" icon="tickets" @click="showYaml('statefulset', row)">YAML</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
      <el-tab-pane label="DaemonSets" name="daemonsets">
        <el-table :data="daemonsets" size="small" v-loading="loading">
          <el-table-column prop="name" label="名称" min-width="160" />
          <el-table-column prop="namespace" label="命名空间" min-width="110" />
          <el-table-column label="就绪" width="100">
            <template #default="{ row }">{{ row.ready }}/{{ row.desired }}</template>
          </el-table-column>
          <el-table-column prop="available" label="可用" width="80" />
          <el-table-column prop="age" label="年龄" width="90" />
          <el-table-column label="操作" width="90" fixed="right">
            <template #default="{ row }">
              <el-button link type="info" icon="tickets" @click="showYaml('daemonset', row)">YAML</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
      <el-tab-pane label="Nodes" name="nodes">
        <el-table :data="nodes" size="small" v-loading="loading">
          <el-table-column prop="name" label="名称" min-width="180" />
          <el-table-column prop="status" label="状态" width="170">
            <template #default="{ row }">
              <el-tag :type="row.status.startsWith('Ready') ? 'success' : 'danger'" size="small">
                {{ row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="version" label="版本" min-width="150" />
          <el-table-column prop="internal" label="InternalIP" min-width="130" />
          <el-table-column prop="age" label="年龄" width="90" />
          <el-table-column label="操作" width="230" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" icon="view" @click="showNodeDetail(row)">详情</el-button>
              <el-button v-if="!row.status.includes('SchedulingDisabled')" link type="warning" icon="lock" @click="onCordon(row, true)">隔离</el-button>
              <el-button v-else link type="success" icon="unlock" @click="onCordon(row, false)">恢复</el-button>
              <el-button link type="danger" icon="switch-button" @click="onDrain(row)">驱逐</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
      <el-tab-pane label="Services" name="services">
        <el-table :data="services" size="small" v-loading="loading">
          <el-table-column prop="name" label="名称" min-width="160" />
          <el-table-column prop="namespace" label="命名空间" min-width="110" />
          <el-table-column prop="type" label="类型" width="100" />
          <el-table-column prop="clusterIp" label="ClusterIP" min-width="130" />
          <el-table-column prop="ports" label="端口" min-width="140" />
          <el-table-column prop="age" label="年龄" width="90" />
        </el-table>
      </el-tab-pane>
      <el-tab-pane label="ConfigMaps" name="configmaps">
        <el-table :data="configMaps" size="small" v-loading="loading">
          <el-table-column prop="name" label="名称" min-width="160" />
          <el-table-column prop="namespace" label="命名空间" min-width="110" />
          <el-table-column prop="dataKeys" label="数据键" min-width="180" />
          <el-table-column prop="age" label="年龄" width="90" />
        </el-table>
      </el-tab-pane>
      <el-tab-pane label="Secrets" name="secrets">
        <el-table :data="secrets" size="small" v-loading="loading">
          <el-table-column prop="name" label="名称" min-width="160" />
          <el-table-column prop="namespace" label="命名空间" min-width="110" />
          <el-table-column prop="type" label="类型" min-width="140" />
          <el-table-column prop="dataKeys" label="数据键" min-width="180" />
          <el-table-column prop="age" label="年龄" width="90" />
        </el-table>
      </el-tab-pane>
      <el-tab-pane label="PVC" name="pvcs">
        <el-table :data="pvcs" size="small" v-loading="loading">
          <el-table-column prop="name" label="名称" min-width="160" />
          <el-table-column prop="namespace" label="命名空间" min-width="110" />
          <el-table-column prop="status" label="状态" width="110">
            <template #default="{ row }">
              <el-tag :type="row.status === 'Bound' ? 'success' : 'warning'" size="small">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="capacity" label="容量" width="100" />
          <el-table-column prop="storageClass" label="StorageClass" min-width="130" />
          <el-table-column prop="volume" label="PV" min-width="150" />
          <el-table-column prop="age" label="年龄" width="90" />
        </el-table>
      </el-tab-pane>
      <el-tab-pane label="Ingress" name="ingresses">
        <el-table :data="ingresses" size="small" v-loading="loading">
          <el-table-column prop="name" label="名称" min-width="150" />
          <el-table-column prop="namespace" label="命名空间" min-width="110" />
          <el-table-column prop="hosts" label="主机" min-width="150" />
          <el-table-column prop="paths" label="路径" min-width="240" show-overflow-tooltip />
          <el-table-column prop="age" label="年龄" width="90" />
        </el-table>
      </el-tab-pane>
      <el-tab-pane label="Events" name="events">
        <el-table :data="events" size="small" v-loading="loading">
          <el-table-column prop="lastTime" label="最近" width="90" />
          <el-table-column prop="type" label="类型" width="90">
            <template #default="{ row }">
              <el-tag :type="row.type === 'Warning' ? 'danger' : 'info'" size="small">{{ row.type || '-' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="reason" label="原因" width="160" />
          <el-table-column prop="object" label="对象" min-width="200" show-overflow-tooltip />
          <el-table-column prop="message" label="消息" min-width="260" show-overflow-tooltip />
          <el-table-column prop="count" label="次数" width="70" />
        </el-table>
      </el-tab-pane>
      <el-tab-pane label="Helm" name="helm">
        <div class="ops-btn-list" style="margin-bottom: 8px">
          <el-button type="primary" size="small" icon="plus" @click="openHelmInstall">安装 release</el-button>
          <el-button size="small" icon="folder" @click="openHelmRepos">仓库管理</el-button>
          <el-button size="small" icon="refresh" @click="loadHelm">刷新</el-button>
        </div>
        <el-table :data="helmReleases" size="small" v-loading="helmLoading">
          <el-table-column prop="name" label="release" min-width="120" />
          <el-table-column prop="namespace" label="命名空间" min-width="100" />
          <el-table-column prop="chart" label="chart" min-width="160" />
          <el-table-column prop="revision" label="版本" width="60" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'deployed' ? 'success' : row.status === 'failed' ? 'danger' : 'warning'" size="small">
                {{ row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="updated" label="更新" width="80" />
          <el-table-column label="操作" width="230" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" icon="view" @click="showHelmDetail(row)">详情</el-button>
              <el-button link type="primary" icon="clock" @click="showHelmHistory(row)">历史</el-button>
              <el-button link type="danger" icon="delete" @click="onHelmUninstall(row)">卸载</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <el-drawer v-model="logsVisible" :title="`日志 · ${logsPod}`" size="60%" append-to-body>
      <pre class="logs-pre">{{ logsText || '（无日志）' }}</pre>
    </el-drawer>

    <el-drawer v-model="nodeDetailVisible" :title="`节点详情 · ${nodeDetail?.name || ''}`" size="55%" append-to-body>
      <div v-if="nodeDetail" v-loading="nodeDetailLoading">
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="名称">{{ nodeDetail.name }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="nodeDetail.status.startsWith('Ready') ? 'success' : 'danger'" size="small">{{ nodeDetail.status }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="角色">{{ nodeDetail.roles }}</el-descriptions-item>
          <el-descriptions-item label="kubelet">{{ nodeDetail.kubelet }}</el-descriptions-item>
          <el-descriptions-item label="InternalIP">{{ nodeDetail.internalIP }}</el-descriptions-item>
          <el-descriptions-item label="架构">{{ nodeDetail.arch }}</el-descriptions-item>
          <el-descriptions-item label="系统镜像" :span="2">{{ nodeDetail.osImage }}</el-descriptions-item>
          <el-descriptions-item label="年龄">{{ nodeDetail.age }}</el-descriptions-item>
        </el-descriptions>
        <h4>可分配 / 容量</h4>
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="Allocatable">{{ nodeDetail.allocatable?.cpu }} CPU · {{ nodeDetail.allocatable?.memory }} 内存 · {{ nodeDetail.allocatable?.pods }} Pods</el-descriptions-item>
          <el-descriptions-item label="Capacity">{{ nodeDetail.capacity?.cpu }} CPU · {{ nodeDetail.capacity?.memory }} 内存 · {{ nodeDetail.capacity?.pods }} Pods</el-descriptions-item>
        </el-descriptions>
        <h4>Conditions</h4>
        <el-table :data="nodeDetail.conditions" size="small" border>
          <el-table-column prop="type" label="类型" width="170" />
          <el-table-column prop="status" label="状态" width="80" />
          <el-table-column prop="reason" label="原因" width="150" />
          <el-table-column prop="message" label="消息" min-width="200" show-overflow-tooltip />
        </el-table>
        <h4>Taints</h4>
        <el-table :data="nodeDetail.taints" size="small" border>
          <el-table-column prop="key" label="Key" min-width="180" />
          <el-table-column prop="value" label="Value" min-width="120" />
          <el-table-column prop="effect" label="Effect" min-width="140" />
        </el-table>
      </div>
    </el-drawer>

    <el-drawer v-model="yamlVisible" :title="`YAML · ${yamlTitle}`" size="55%" append-to-body>
      <div v-loading="yamlLoading">
        <div class="yaml-toolbar">
          <el-button v-if="!yamlEditing" type="primary" size="small" icon="edit" @click="startEditYaml">编辑</el-button>
          <template v-else>
            <el-button type="warning" size="small" icon="view" :loading="yamlDiffing" @click="previewYamlChange">预览变更</el-button>
            <el-button size="small" @click="cancelEditYaml">取消</el-button>
          </template>
          <span v-if="yamlEditing" class="yaml-tip">编辑后先预览变更，确认 diff 再下发（乐观锁保护）</span>
        </div>
        <pre v-if="!yamlEditing" class="logs-pre">{{ yamlText || '（空）' }}</pre>
        <el-input
          v-else
          v-model="yamlDraft"
          type="textarea"
          :rows="26"
          class="yaml-editor"
          spellcheck="false"
        />
      </div>
    </el-drawer>

    <el-dialog v-model="yamlDiffVisible" title="变更预览（服务端 dry-run diff）" width="70%" append-to-body>
      <div v-loading="yamlDiffing" class="yaml-diff-box">
        <div v-for="(line, i) in yamlDiffLines" :key="i" class="yaml-diff-line" :class="diffLineClass(line)">{{ line }}</div>
      </div>
      <template #footer>
        <el-button @click="yamlDiffVisible = false">取 消</el-button>
        <el-button type="danger" :loading="yamlApplying" @click="confirmApplyYaml">确认下发</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="helmInstallVisible" title="安装/升级 Helm release" width="580px" append-to-body>
      <el-form label-width="90px">
        <el-form-item label="chart 来源">
          <el-radio-group v-model="helmInstallForm.mode">
            <el-radio value="upload">上传 .tgz</el-radio>
            <el-radio value="repo">仓库引用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="命名空间" required>
          <el-input v-model="helmInstallForm.namespace" placeholder="如 default" />
        </el-form-item>
        <el-form-item label="release 名" required>
          <el-input v-model="helmInstallForm.releaseName" placeholder="如 my-app" />
        </el-form-item>
        <template v-if="helmInstallForm.mode === 'upload'">
          <el-form-item label="chart 包" required>
            <input type="file" accept=".tgz" @change="(e) => (helmInstallForm.file = e.target.files[0])" />
          </el-form-item>
        </template>
        <template v-else>
          <el-form-item label="仓库" required>
            <el-select v-model="helmInstallForm.repoId" style="width: 100%" placeholder="选择已登记仓库">
              <el-option v-for="r in helmRepos" :key="r.ID" :label="`${r.name}（${r.url}）`" :value="r.ID" />
            </el-select>
          </el-form-item>
          <el-form-item label="chart 名" required>
            <el-input v-model="helmInstallForm.chart" placeholder="如 nginx-ingress" />
          </el-form-item>
          <el-form-item label="版本">
            <el-input v-model="helmInstallForm.version" placeholder="留空=最新" />
          </el-form-item>
        </template>
        <el-form-item label="values">
          <el-input v-model="helmInstallForm.values" type="textarea" :rows="6" placeholder="values YAML 覆盖（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="helmInstallVisible = false">取 消</el-button>
        <el-button type="primary" :loading="helmInstalling" @click="submitHelmInstall">安装/升级</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="helmRepoVisible" title="chart 仓库管理" width="640px" append-to-body>
      <div class="ops-btn-list" style="margin-bottom: 8px">
        <el-input v-model="helmRepoForm.name" placeholder="名称" style="width: 140px" />
        <el-input v-model="helmRepoForm.url" placeholder="http(s)://repo/index.yaml" style="width: 280px" />
        <el-button type="primary" icon="plus" @click="submitHelmRepo">登记</el-button>
      </div>
      <el-table :data="helmRepos" size="small" border>
        <el-table-column prop="name" label="名称" width="130" />
        <el-table-column prop="url" label="地址" min-width="260" show-overflow-tooltip />
        <el-table-column prop="remark" label="备注" min-width="120" />
        <el-table-column label="操作" width="80">
          <template #default="{ row }">
            <el-button link type="danger" icon="delete" @click="onDeleteHelmRepo(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <el-drawer v-model="helmDetailVisible" :title="`release 详情 · ${helmDetailName}`" size="60%" append-to-body>
      <div v-loading="helmDetailLoading">
        <el-descriptions v-if="helmDetail" :column="2" border size="small" style="margin-bottom: 10px">
          <el-descriptions-item label="chart">{{ helmDetail.chart }}</el-descriptions-item>
          <el-descriptions-item label="版本">{{ helmDetail.revision }}（{{ helmDetail.status }}）</el-descriptions-item>
        </el-descriptions>
        <h4>values（当前生效）</h4>
        <pre class="logs-pre" style="max-height: 30vh">{{ helmDetail?.values || '（空）' }}</pre>
        <h4>渲染 manifest</h4>
        <pre class="logs-pre" style="max-height: 40vh">{{ helmDetail?.manifest || '（空）' }}</pre>
      </div>
    </el-drawer>

    <el-dialog v-model="helmHistoryVisible" :title="`历史 · ${helmHistoryName}`" width="560px" append-to-body>
      <el-table :data="helmHistory" size="small" border v-loading="helmHistoryLoading">
        <el-table-column prop="revision" label="版本" width="70" />
        <el-table-column prop="status" label="状态" width="110" />
        <el-table-column prop="chart" label="chart" min-width="160" />
        <el-table-column prop="updated" label="更新" width="90" />
        <el-table-column label="操作" width="90">
          <template #default="{ row }">
            <el-button link type="warning" icon="refresh-left" @click="onHelmRollback(row)">回滚到此</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <el-drawer v-model="shellVisible" :title="`Pod 终端`" size="65%" append-to-body destroy-on-close>
      <PodShell
        v-if="shellVisible && shellTarget.clusterId"
        :cluster-id="shellTarget.clusterId"
        :namespace="shellTarget.namespace"
        :pod="shellTarget.pod"
      />
    </el-drawer>

    <el-drawer v-model="podDetailVisible" :title="`Pod 详情 · ${podDetail?.name || ''}`" size="60%" append-to-body>
      <div v-if="podDetail" v-loading="podDetailLoading">
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="命名空间">{{ podDetail.namespace }}</el-descriptions-item>
          <el-descriptions-item label="阶段">
            <el-tag :type="podDetail.phase === 'Running' ? 'success' : 'warning'" size="small">{{ podDetail.phase }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="节点">{{ podDetail.node || '-' }}</el-descriptions-item>
          <el-descriptions-item label="PodIP">{{ podDetail.podIP || '-' }}</el-descriptions-item>
          <el-descriptions-item label="HostIP">{{ podDetail.hostIP || '-' }}</el-descriptions-item>
          <el-descriptions-item label="QoS">{{ podDetail.qoS || '-' }}</el-descriptions-item>
          <el-descriptions-item label="重启总数">{{ podDetail.restarts }}</el-descriptions-item>
          <el-descriptions-item label="年龄">{{ podDetail.age }}</el-descriptions-item>
        </el-descriptions>
        <h4>容器</h4>
        <el-table :data="podDetail.containers" size="small" border>
          <el-table-column prop="name" label="名称" min-width="130" />
          <el-table-column prop="image" label="镜像" min-width="200" show-overflow-tooltip />
          <el-table-column label="就绪" width="70">
            <template #default="{ row }">
              <el-tag :type="row.ready ? 'success' : 'danger'" size="small">{{ row.ready ? '是' : '否' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="restarts" label="重启" width="60" />
          <el-table-column prop="state" label="状态" width="120">
            <template #default="{ row }">
              <el-tag :type="row.state === 'Running' ? 'success' : row.state === 'Waiting' ? 'warning' : 'info'" size="small">
                {{ row.state }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="reason" label="原因" min-width="140" show-overflow-tooltip />
        </el-table>
        <h4>Conditions</h4>
        <el-table :data="podDetail.conditions" size="small" border>
          <el-table-column prop="type" label="类型" width="160" />
          <el-table-column prop="status" label="状态" width="80" />
          <el-table-column prop="reason" label="原因" min-width="160" show-overflow-tooltip />
        </el-table>
        <h4>Events</h4>
        <el-table :data="podDetail.events" size="small" border>
          <el-table-column prop="lastTime" label="最近" width="90" />
          <el-table-column prop="type" label="类型" width="80" />
          <el-table-column prop="reason" label="原因" width="150" />
          <el-table-column prop="message" label="消息" min-width="220" show-overflow-tooltip />
          <el-table-column prop="count" label="次数" width="60" />
        </el-table>
      </div>
    </el-drawer>
  </el-drawer>
</template>

<script setup>
  import {
    getK8sClusterList, getK8sClusterPodList, getK8sClusterPodLogs,
    getK8sClusterDeploymentList, getK8sClusterNodeList,
    getK8sServiceList, getK8sConfigMapList, getK8sSecretList,
    getK8sNodeDetail, cordonK8sNode, drainK8sNode,
    getK8sStatefulSetList, getK8sDaemonSetList, getK8sWorkloadYaml,
    previewK8sWorkloadYaml, applyK8sWorkloadYaml,
    getK8sPodDetail, deleteK8sPod,
    getK8sPvcList, getK8sIngressList, getK8sEventList,
    getHelmList, getHelmHistory, installHelmRelease, uninstallHelmRelease, rollbackHelmRelease,
    installHelmFromRepo, getHelmReleaseDetail, getHelmRepoList, createHelmRepo, deleteHelmRepo,
    getNsVisibility
  } from '@/plugin/k8s/api/k8sResource'
  import { ElMessage, ElMessageBox } from 'element-plus'
  import { ref, watch } from 'vue'
  import PodShell from '@/plugin/k8s/components/PodShell.vue'

  const visible = defineModel('visible', { type: Boolean })
  const props = defineProps({
    clusterId: { type: Number, required: true },
    clusterName: { type: String, default: '' }
  })

  const activeTab = ref('pods')
  const loading = ref(false)
  // 命名空间过滤（普通用户下拉仅列授权项——后端 ns-visibility 按当前用户返回）
  const nsFilter = ref('')
  const nsOptions = ref([])
  const pods = ref([])
  const deployments = ref([])
  const statefulsets = ref([])
  const daemonsets = ref([])
  const nodes = ref([])
  const services = ref([])
  const configMaps = ref([])
  const secrets = ref([])
  const pvcs = ref([])
  const ingresses = ref([])
  const events = ref([])

  // Helm release 管理
  const helmReleases = ref([])
  const helmLoading = ref(false)
  const helmInstallVisible = ref(false)
  const helmInstalling = ref(false)
  const emptyHelmInstallForm = () => ({
    mode: 'upload', namespace: 'default', releaseName: '', values: '',
    file: null, repoId: null, chart: '', version: ''
  })
  const helmInstallForm = ref(emptyHelmInstallForm())
  const helmHistoryVisible = ref(false)
  const helmHistoryLoading = ref(false)
  const helmHistory = ref([])
  const helmHistoryName = ref('')
  const helmRepoVisible = ref(false)
  const helmRepos = ref([])
  const helmRepoForm = ref({ name: '', url: '' })
  const helmDetailVisible = ref(false)
  const helmDetailLoading = ref(false)
  const helmDetail = ref(null)
  const helmDetailName = ref('')

  const loadHelmRepos = async () => {
    const res = await getHelmRepoList()
    if (res.code === 0) helmRepos.value = res.data || []
  }

  const openHelmRepos = () => {
    helmRepoVisible.value = true
    loadHelmRepos()
  }

  const openHelmInstall = () => {
    helmInstallForm.value = emptyHelmInstallForm()
    helmInstallVisible.value = true
    loadHelmRepos()
  }

  const submitHelmRepo = async () => {
    if (!helmRepoForm.value.name || !helmRepoForm.value.url) {
      ElMessage.warning('名称与地址必填')
      return
    }
    const res = await createHelmRepo(helmRepoForm.value)
    if (res.code === 0) {
      ElMessage.success('登记成功')
      helmRepoForm.value = { name: '', url: '' }
      loadHelmRepos()
    }
  }

  const onDeleteHelmRepo = (row) => {
    ElMessageBox.confirm(`确定删除仓库「${row.name}」登记吗？`, '提示', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning'
    }).then(async () => {
      const res = await deleteHelmRepo(row.ID)
      if (res.code === 0) {
        ElMessage.success('删除成功')
        loadHelmRepos()
      }
    })
  }

  const showHelmDetail = async (row) => {
    helmDetailName.value = `${row.namespace}/${row.name}`
    helmDetailVisible.value = true
    helmDetailLoading.value = true
    helmDetail.value = null
    try {
      const res = await getHelmReleaseDetail({
        clusterId: props.clusterId, namespace: row.namespace, name: row.name
      })
      if (res.code === 0) helmDetail.value = res.data
      else ElMessage.error(res.msg || '详情获取失败')
    } finally {
      helmDetailLoading.value = false
    }
  }

  const loadHelm = async () => {
    helmLoading.value = true
    try {
      const res = await getHelmList({ clusterId: props.clusterId })
      if (res.code === 0) helmReleases.value = res.data || []
    } finally {
      helmLoading.value = false
    }
  }

  const submitHelmInstall = async () => {
    const f = helmInstallForm.value
    if (!f.namespace || !f.releaseName) {
      ElMessage.warning('命名空间/release 名必填')
      return
    }
    helmInstalling.value = true
    try {
      let res
      if (f.mode === 'upload') {
        if (!f.file) {
          ElMessage.warning('chart 包必填')
          return
        }
        const form = new FormData()
        form.append('namespace', f.namespace)
        form.append('releaseName', f.releaseName)
        form.append('values', f.values || '')
        form.append('chart', f.file)
        res = await installHelmRelease(props.clusterId, form)
      } else {
        if (!f.repoId || !f.chart) {
          ElMessage.warning('仓库与 chart 名必填')
          return
        }
        res = await installHelmFromRepo(props.clusterId, {
          repoId: f.repoId,
          namespace: f.namespace,
          releaseName: f.releaseName,
          chart: f.chart,
          version: f.version || '',
          values: f.values || ''
        })
      }
      if (res.code === 0) {
        ElMessage.success(`安装/升级成功：${res.data?.name || f.releaseName} rev.${res.data?.revision ?? '-'}`)
        helmInstallVisible.value = false
        helmInstallForm.value = emptyHelmInstallForm()
        loadHelm()
      }
    } finally {
      helmInstalling.value = false
    }
  }

  const showHelmHistory = async (row) => {
    helmHistoryName.value = `${row.namespace}/${row.name}`
    helmHistoryVisible.value = true
    helmHistoryLoading.value = true
    helmHistory.value = []
    try {
      const res = await getHelmHistory({
        clusterId: props.clusterId, namespace: row.namespace, name: row.name
      })
      if (res.code === 0) helmHistory.value = res.data || []
    } finally {
      helmHistoryLoading.value = false
    }
  }

  const onHelmRollback = (row) => {
    const target = helmHistoryName.value.split('/')
    ElMessageBox.confirm(`确定回滚到版本 ${row.revision} 吗？`, '提示', {
      confirmButtonText: '回滚', cancelButtonText: '取消', type: 'warning'
    }).then(async () => {
      const res = await rollbackHelmRelease(props.clusterId, target[0], target[1], row.revision)
      if (res.code === 0) {
        ElMessage.success('回滚成功')
        helmHistoryVisible.value = false
        loadHelm()
      }
    })
  }

  const onHelmUninstall = (row) => {
    ElMessageBox.confirm(
      `确定卸载 release「${row.namespace}/${row.name}」吗？其全部资源将被删除。`,
      '危险操作',
      { confirmButtonText: '卸载', cancelButtonText: '取消', type: 'error' }
    ).then(async () => {
      const res = await uninstallHelmRelease(props.clusterId, row.namespace, row.name)
      if (res.code === 0) {
        ElMessage.success('卸载成功')
        loadHelm()
      }
    })
  }
  const logsVisible = ref(false)
  const logsPod = ref('')
  const logsText = ref('')

  const loadAll = async () => {
    loading.value = true
    const ns = nsFilter.value || ''
    try {
      const [p, d, sts, ds, n, sv, cm, sec, pvc, ing, ev] = await Promise.all([
        getK8sClusterPodList({ clusterId: props.clusterId, namespace: ns }),
        getK8sClusterDeploymentList({ clusterId: props.clusterId, namespace: ns }),
        getK8sStatefulSetList({ clusterId: props.clusterId, namespace: ns }),
        getK8sDaemonSetList({ clusterId: props.clusterId, namespace: ns }),
        getK8sClusterNodeList({ clusterId: props.clusterId }),
        getK8sServiceList({ clusterId: props.clusterId, namespace: ns }),
        getK8sConfigMapList({ clusterId: props.clusterId, namespace: ns }),
        getK8sSecretList({ clusterId: props.clusterId, namespace: ns }),
        getK8sPvcList({ clusterId: props.clusterId, namespace: ns }),
        getK8sIngressList({ clusterId: props.clusterId, namespace: ns }),
        getK8sEventList({ clusterId: props.clusterId, namespace: ns })
      ])
      pods.value = p.code === 0 ? p.data : []
      deployments.value = d.code === 0 ? d.data : []
      statefulsets.value = sts.code === 0 ? sts.data : []
      daemonsets.value = ds.code === 0 ? ds.data : []
      nodes.value = n.code === 0 ? n.data : []
      services.value = sv.code === 0 ? sv.data : []
      configMaps.value = cm.code === 0 ? cm.data : []
      secrets.value = sec.code === 0 ? sec.data : []
      pvcs.value = pvc.code === 0 ? pvc.data : []
      ingresses.value = ing.code === 0 ? ing.data : []
      events.value = ev.code === 0 ? ev.data : []
      loadHelm()
    } finally {
      loading.value = false
    }
  }

  const showLogs = async (pod) => {
    logsPod.value = pod.name
    const res = await getK8sClusterPodLogs({
      clusterId: props.clusterId, namespace: pod.namespace,
      pod: pod.name, tailLines: 500
    })
    logsText.value = res.code === 0 ? res.data : `获取失败：${res.msg}`
    logsVisible.value = true
  }

  // 节点详情 / 隔离 / 驱逐
  const nodeDetailVisible = ref(false)
  const nodeDetailLoading = ref(false)
  const nodeDetail = ref(null)

  const showNodeDetail = async (row) => {
    nodeDetailVisible.value = true
    nodeDetailLoading.value = true
    nodeDetail.value = null
    try {
      const res = await getK8sNodeDetail({ clusterId: props.clusterId, name: row.name })
      if (res.code === 0) {
        nodeDetail.value = res.data
      } else {
        ElMessage.error(res.msg || '节点详情获取失败')
        nodeDetailVisible.value = false
      }
    } finally {
      nodeDetailLoading.value = false
    }
  }

  const refreshNodes = async () => {
    const n = await getK8sClusterNodeList({ clusterId: props.clusterId })
    if (n.code === 0) nodes.value = n.data
  }

  const onCordon = (row, cordon) => {
    ElMessageBox.confirm(
      cordon ? `确定隔离节点「${row.name}」吗？隔离后新 Pod 不会调度到该节点。` : `确定恢复节点「${row.name}」调度吗？`,
      '提示',
      { confirmButtonText: cordon ? '隔离' : '恢复', cancelButtonText: '取消', type: 'warning' }
    ).then(async () => {
      const res = await cordonK8sNode(props.clusterId, row.name, cordon)
      if (res.code === 0) {
        ElMessage.success(res.msg || '操作成功')
        refreshNodes()
      }
    })
  }

  const onDrain = (row) => {
    ElMessageBox.confirm(
      `确定驱逐节点「${row.name}」上的全部可驱逐 Pod 吗？将先隔离节点，再逐个提交 Eviction（DaemonSet 与静态 Pod 跳过）。`,
      '危险操作',
      { confirmButtonText: '驱逐', cancelButtonText: '取消', type: 'error' }
    ).then(async () => {
      const res = await drainK8sNode(props.clusterId, row.name)
      if (res.code === 0) {
        const d = res.data || {}
        ElMessage.success(`驱逐已提交：成功 ${d.evicted?.length || 0}，跳过 ${d.skipped?.length || 0}，失败 ${d.failed?.length || 0}`)
        refreshNodes()
      }
    })
  }

  // 工作负载 YAML
  const yamlVisible = ref(false)
  const yamlLoading = ref(false)
  const yamlTitle = ref('')
  const yamlText = ref('')
  const yamlKind = ref('')

  // YAML 编辑下发（diff 预览 → 确认 apply）
  const yamlEditing = ref(false)
  const yamlDraft = ref('')
  const yamlDiffVisible = ref(false)
  const yamlDiffing = ref(false)
  const yamlApplying = ref(false)
  const yamlDiffLines = ref([])
  const yamlTarget = ref({ namespace: '', name: '' })

  const diffLineClass = (line) => {
    if (line.startsWith('+')) return 'diff-add'
    if (line.startsWith('-')) return 'diff-del'
    return 'diff-same'
  }

  const startEditYaml = () => {
    yamlDraft.value = yamlText.value
    yamlEditing.value = true
  }

  const cancelEditYaml = () => {
    yamlEditing.value = false
    yamlDraft.value = ''
  }

  const previewYamlChange = async () => {
    yamlDiffing.value = true
    try {
      const res = await previewK8sWorkloadYaml(props.clusterId, {
        kind: yamlKind.value,
        namespace: yamlTarget.value.namespace,
        name: yamlTarget.value.name,
        yaml: yamlDraft.value
      })
      if (res.code !== 0) {
        ElMessage.error(res.msg || '预览失败')
        return
      }
      const diff = String(res.data || '')
      if (diff.includes('（无实质变更）')) {
        ElMessage.info('无实质变更，无需下发')
        return
      }
      yamlDiffLines.value = diff.split('\n')
      yamlDiffVisible.value = true
    } finally {
      yamlDiffing.value = false
    }
  }

  const confirmApplyYaml = async () => {
    yamlApplying.value = true
    try {
      const res = await applyK8sWorkloadYaml(props.clusterId, {
        kind: yamlKind.value,
        namespace: yamlTarget.value.namespace,
        name: yamlTarget.value.name,
        yaml: yamlDraft.value
      })
      if (res.code === 0) {
        ElMessage.success(res.msg || 'YAML 已下发')
        yamlDiffVisible.value = false
        yamlEditing.value = false
        yamlVisible.value = false
        loadAll()
      }
    } finally {
      yamlApplying.value = false
    }
  }

  // Pod 终端
  const shellVisible = ref(false)
  const shellTarget = ref({ clusterId: 0, namespace: '', pod: '' })
  const openShell = (row) => {
    shellTarget.value = {
      clusterId: props.clusterId, namespace: row.namespace, pod: row.name
    }
    shellVisible.value = true
  }

  const showYaml = async (kind, row) => {
    yamlTitle.value = `${row.namespace}/${row.name}`
    yamlKind.value = kind
    yamlTarget.value = { namespace: row.namespace, name: row.name }
    yamlEditing.value = false
    yamlVisible.value = true
    yamlLoading.value = true
    yamlText.value = ''
    try {
      const res = await getK8sWorkloadYaml({
        clusterId: props.clusterId, kind, namespace: row.namespace, name: row.name
      })
      yamlText.value = res.code === 0 ? res.data : `获取失败：${res.msg}`
    } finally {
      yamlLoading.value = false
    }
  }

  // Pod 详情 / 删除
  const podDetailVisible = ref(false)
  const podDetailLoading = ref(false)
  const podDetail = ref(null)

  const showPodDetail = async (row) => {
    podDetailVisible.value = true
    podDetailLoading.value = true
    podDetail.value = null
    try {
      const res = await getK8sPodDetail({
        clusterId: props.clusterId, namespace: row.namespace, name: row.name
      })
      if (res.code === 0) {
        podDetail.value = res.data
      } else {
        ElMessage.error(res.msg || 'Pod 详情获取失败')
        podDetailVisible.value = false
      }
    } finally {
      podDetailLoading.value = false
    }
  }

  const refreshPods = async () => {
    const p = await getK8sClusterPodList({ clusterId: props.clusterId })
    if (p.code === 0) pods.value = p.data
  }

  const onDeletePod = (row) => {
    ElMessageBox.confirm(
      `确定删除 Pod「${row.namespace}/${row.name}」吗？无控制器的 Pod 将永久移除，有控制器的会被重建。`,
      '危险操作',
      { confirmButtonText: '删除', cancelButtonText: '取消', type: 'error' }
    ).then(async () => {
      const res = await deleteK8sPod(props.clusterId, row.namespace, row.name)
      if (res.code === 0) {
        ElMessage.success(res.msg || '删除已提交')
        refreshPods()
      }
    })
  }

  watch(visible, (v) => {
    if (v) {
      loadAll()
      getNsVisibility({ clusterId: props.clusterId }).then((res) => {
        if (res.code === 0) {
          nsOptions.value = (res.data || []).filter((n) => n.granted)
        }
      })
    }
  })
</script>

<style scoped>
.logs-pre {
  white-space: pre-wrap;
  word-break: break-all;
  font-size: 12px;
  max-height: 65vh;
  overflow: auto;
}
.yaml-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
}
.yaml-tip {
  font-size: 12px;
  color: #909399;
}
.yaml-editor :deep(textarea) {
  font-family: Consolas, Monaco, 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.5;
}
.yaml-diff-box {
  max-height: 55vh;
  overflow: auto;
  background: #0b1021;
  border-radius: 4px;
  padding: 8px 0;
  font-family: Consolas, Monaco, 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.6;
}
.yaml-diff-line {
  white-space: pre-wrap;
  word-break: break-all;
  padding: 0 12px;
}
.diff-add {
  color: #4ade80;
  background: rgba(74, 222, 128, 0.08);
}
.diff-del {
  color: #f87171;
  background: rgba(248, 113, 113, 0.08);
}
.diff-same {
  color: #94a3b8;
}
</style>
