<template>
  <div>
    <clusterbar :titleName="titleName" :nsFunc="nsSearch" :nameFunc="nameSearch" :delFunc="delFunc"/>
      <!-- :createFunc="createFunc" createDisplay="创建"/> -->
    <div class="dashboard-container">
      <!-- <div class="dashboard-text"></div> -->
      <el-table
        ref="multipleTable"
        :data="statefulsets"
        class="table-fix"
        tooltip-effect="dark"
        :max-height="maxHeight"
        style="width: 100%"
        v-loading="loading"
        :cell-style="cellStyle"
        :default-sort = "{prop: 'name'}"
        @selection-change="handleSelectionChange"
        row-key="uid"
        >
        <el-table-column
          type="selection"
          width="45">
        </el-table-column>
        <el-table-column
          prop="name"
          label="名称"
          min-width="100"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <span class="name-class" v-on:click="nameClick(scope.row.namespace, scope.row.name)">
              {{ scope.row.name }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="namespace"
          label="命名空间"
          min-width="60"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="ready_replicas"
          label="Pods"
          min-width="40"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <span>
              {{ scope.row.ready_replicas }}/{{ scope.row.status_replicas }}
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="replicas"
          label="副本"
          min-width="30"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="strategy"
          label="更新策略"
          min-width="55"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="conditions"
          label="状态"
          min-width="45"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <template v-if="scope.row.conditions && scope.row.conditions.length > 0">
              <span v-for="c in scope.row.conditions" :key="c">
                {{ c }}
              </span>
            </template>
            <span v-else>—</span>
          </template>
        </el-table-column>
        <el-table-column
          prop="created"
          label="创建时间"
          min-width="55"
          show-overflow-tooltip>
          <template slot-scope="scope">
            {{ $dateFormat(scope.row.created) }}
          </template>
        </el-table-column>
        <el-table-column
          label=""
          show-overflow-tooltip
          width="45">
          <template slot-scope="scope">
            <el-dropdown size="medium" >
              <el-link :underline="false"><svg-icon style="width: 1.3em; height: 1.3em;" icon-class="operate" /></el-link>
              <el-dropdown-menu slot="dropdown">
                <el-dropdown-item @click.native.prevent="nameClick(scope.row.namespace, scope.row.name)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="detail" />
                  <span style="margin-left: 5px;">详情</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$editorRole()" 
                  @click.native.prevent="update_replicas_statefulset = scope.row; update_replicas = scope.row.replicas; replicaDialog = true;">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="scale" />
                  <span style="margin-left: 5px;">副本</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$editorRole()" @click.native.prevent="getStatefulSetYaml(scope.row.namespace, scope.row.name)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="edit" />
                  <span style="margin-left: 5px;">修改</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$editorRole()" @click.native.prevent="deleteStatefulSets([{namespace: scope.row.namespace, name: scope.row.name}])">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="delete" />
                  <span style="margin-left: 5px;">删除</span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>
    </div>
    <el-dialog title="编辑" :visible.sync="yamlDialog" :close-on-click-modal="false" width="60%" top="55px" v-loading="yamlLoading">
      <yaml v-if="yamlDialog" v-model="yamlValue" :loading="yamlLoading"></yaml>
      <span slot="footer" class="dialog-footer">
        <el-button plain @click="yamlDialog = false" size="small">取 消</el-button>
        <el-button plain @click="updateStatefulSet()" size="small">确 定</el-button>
      </span>
    </el-dialog>
    <el-dialog title="扩缩容" :visible.sync="replicaDialog" :close-on-click-modal="false" width="380px" top="14%" class="replicaDialog" >
      <div slot="title">
        <span style="line-height: 24px; font-size: 18px; color: #303133;">扩缩容:</span>
        <span style="line-height: 24px; font-size: 15px; color: #606266;">{{ update_replicas_statefulset ? update_replicas_statefulset.name : '' }}</span>
      </div>
      <el-form ref="form" label-position="left" label-width="100px">
        <el-form-item label="当前副本数">
          <label>{{ update_replicas_statefulset ? update_replicas_statefulset.replicas : 0 }}</label>
        </el-form-item>
      </el-form>
      <el-form ref="form" label-position="left" label-width="100px">
        <el-form-item label="期望副本数">
          <el-input-number size="medium" v-model="update_replicas" :min="1" :max="100"></el-input-number>
        </el-form-item>
      </el-form>
      <span slot="footer" class="dialog-footer">
        <el-button plain @click="replicaDialog = false" size="small">取 消</el-button>
        <el-button plain @click="updateStatefulSetObj({statefulset: update_replicas_statefulset, replicas: update_replicas})" size="small">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { ResType, listResource, watchResource, getResource, delResource, updateResource, patchResource } from '@/api/cluster/resource'
import { Message } from 'element-ui'
import { Yaml } from '@/views/components'

export default {
  name: 'StatefulSet',
  components: {
    Clusterbar,
    Yaml
  },
  data() {
      return {
        replicaDialog: false,
        yamlDialog: false,
        yamlNamespace: "",
        yamlName: "",
        yamlValue: "",
        yamlLoading: true,
        cellStyle: {border: 0},
        titleName: ["StatefulSets"],
        maxHeight: window.innerHeight - this.$contentHeight,
        loading: true,
        originStatefulSets: [],
        search_ns: [],
        search_name: '',
        delFunc: undefined,
        delStatefulSets: [],
        update_replicas: 0,
        update_replicas_statefulset: null,
        clusterSSE: undefined
      }
  },
  created() {
    this.fetchData()
  },
  mounted() {
    const that = this
    window.onresize = () => {
      return (() => {
        let heightStyle = window.innerHeight - this.$contentHeight
        // console.log(heightStyle)
        that.maxHeight = heightStyle
      })()
    }
  },
  beforeDestroy() {
    if(this.clusterSSE) this.clusterSSE.disconnect()
  },
  watch: {
    cluster: function(obj) {
      this.fetchData()
    }
  },
  computed: {
    statefulsets: function() {
      let dlist = []
      for (let p of this.originStatefulSets) {
        if (this.search_ns.length > 0 && this.search_ns.indexOf(p.namespace) < 0) continue
        if (this.search_name && !p.name.includes(this.search_name)) continue
        if (p.conditions && p.conditions.length > 0) {
          p.conditions.sort()
        } else {
          p.conditions = []
        }
        dlist.push(p)
      }
      return dlist
    },
    cluster() {
      return this.$store.state.cluster
    }
  },
  methods: {
    fetchData: function() {
      this.loading = true
      this.originStatefulSets = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listResource(cluster, ResType.Statefulset).then(response => {
          this.loading = false
          this.originStatefulSets = response.data || []
          if(!this.clusterSSE) this.fetchSSE()
        }).catch(() => {
          this.loading = false
        })
      } else {
        this.loading = false
        Message.error("获取集群异常，请刷新重试")
      }
    },
    fetchSSE() {
      if(!this.clusterSSE) {
        this.clusterSSE = watchResource(this.$sse, this.cluster, ResType.Statefulset, this.sseWatch, {process: true})
      }
    },
    sseWatch(newObj) {
      if (newObj) {
        let newUid = newObj.resource.uid
        let newRv = newObj.resource.resource_version
        if (newObj.event === 'add') {
          for(let i in this.originStatefulSets) {
            let d = this.originStatefulSets[i]
            if (d.uid === newUid) return
          }
          this.originStatefulSets.push(newObj.resource)
        } else if (newObj.event === 'update') {
          for (let i in this.originStatefulSets) {
            let d = this.originStatefulSets[i]
            if (d.uid === newUid) {
              if (d.resource_version < newRv){
                let newDp = newObj.resource
                this.$set(this.originStatefulSets, i, newDp)
              }
              break
            }
          }
        } else if (newObj.event === 'delete') {
          this.originStatefulSets = this.originStatefulSets.filter(( { uid } ) => uid !== newUid)
        }
      }
    },
    nsSearch: function(vals) {
      this.search_ns = []
      for(let ns of vals) {
        this.search_ns.push(ns)
      }
    },
    nameSearch: function(val) {
      this.search_name = val
    },
    nameClick: function(namespace, name) {
      this.$router.push({name: 'statefulsetDetail', params: {namespace: namespace, statefulsetName: name}})
    },
    getStatefulSetYaml: function(namespace, name) {
      this.yamlNamespace = ""
      this.yamlName = ""
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if (!namespace) {
        Message.error("获取命名空间参数异常，请刷新重试")
        return
      }
      if (!name) {
        Message.error("获取StatefulSet名称参数异常，请刷新重试")
        return
      }
      this.yamlValue = ""
      this.yamlDialog = true
      this.yamlLoading = true
      getResource(cluster, ResType.Statefulset, namespace, name, "yaml").then(response => {
        this.yamlLoading = false
        this.yamlValue = response.data
        this.yamlNamespace = namespace
        this.yamlName = name
      }).catch(() => {
        this.yamlLoading = false
      })
    },
    deleteStatefulSets: function(statefulsets) {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if ( statefulsets.length <= 0 ){
        Message.error("请选择要删除的StatefulSets")
        return
      }
      let params = {
        resources: statefulsets
      }
      delResource(cluster, ResType.Statefulset, params).then(() => {
        Message.success("删除成功")
      }).catch(() => {
        // console.log(e)
      })
    },
    updateStatefulSet: function() {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if (!this.yamlNamespace) {
        Message.error("获取命名空间参数异常，请刷新重试")
        return
      }
      if (!this.yamlName) {
        Message.error("获取StatefulSet参数异常，请刷新重试")
        return
      }
      this.yamlLoading = true
      updateResource(cluster, ResType.Statefulset, this.yamlNamespace, this.yamlName, this.yamlValue).then(() => {
        Message.success("更新成功")
        this.yamlLoading = false
        this.yamlDialog = false
      }).catch(() => {
        // console.log(e) 
      })
    },
    updateStatefulSetObj: function(update_obj) {
      console.log(update_obj)
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if (!update_obj || !update_obj.statefulset) {
        Message.error("获取更新参数异常，请刷新重试")
        return
      }
      let statefulset = update_obj.statefulset
      if (!statefulset.namespace) {
        Message.error("获取命名空间参数异常，请刷新重试")
        return
      }
      if (!statefulset.name) {
        Message.error("获取StatefulSet参数异常，请刷新重试")
        return
      }
      let update_params = {
        name: statefulset.name,
        namespace: statefulset.namespace,
      }
      if (update_obj.replicas) {
        if (update_obj.replicas < 1) {
          Message.error("副本数不能小于1，请重新输入")
          return
        }
        if (parseInt(update_obj.replicas) === parseInt(statefulset.replicas)) {
          Message.error("副本数与当前值相同，请重新输入")
          return
        }
        update_params["data"] = {spec: {replicas: update_obj.replicas}}
      }
      if (Object.keys(update_params).length === 0) {
        Message.error("更新参数为空")
        return
      }
      this.yamlLoading = true
      patchResource(cluster, ResType.Statefulset, update_params).then(() => {
        Message.success("更新成功")
        this.yamlLoading = false
        this.replicaDialog = false;
      }).catch(() => {
        // console.log(e) 
      })
    },
    _delStatefulSetsFunc: function() {
      if (this.delStatefulSets.length > 0){
        let delStatefulSets = []
        for (var p of this.delStatefulSets) {
          delStatefulSets.push({namespace: p.namespace, name: p.name})
        }
        this.deleteStatefulSets(delStatefulSets)
      }
    },
    handleSelectionChange(val) {
      this.delStatefulSets = val;
      if (val.length > 0){
        this.delFunc = this._delStatefulSetsFunc
      } else {
        this.delFunc = undefined
      }
    },
    createFunc() {
      this.$router.push({name: 'statefulsetCreate'})
    }
  }
}
</script>

<style lang="scss" scoped>
.dashboard {
  &-container {
    margin: 10px 30px;
  }
  &-text {
    font-size: 30px;
    line-height: 46px;
  }

  .table-fix {
    height: calc(100% - 100px);
  }
}

.name-class {
  cursor: pointer;
}
.name-class:hover {
  color: #409EFF;
}

.scrollbar-wrapper {
  overflow-x: hidden !important;
}
.el-scrollbar__bar.is-vertical {
  right: 0px;
}

.el-scrollbar {
  height: 100%;
}

.running-class {
  color: #67C23A;
}

.terminate-class {
  color: #909399;
}

.waiting-class {
  color: #E6A23C;
}
</style>

<style lang="scss">
.el-dialog__body {
  padding-top: 5px;
  padding-bottom: 5px;
}
.replicaDialog {
  .el-form-item {
    margin-bottom: 10px;
  }
  .el-dialog--center .el-dialog__body {
    padding: 5px 25px;
  }
}
</style>
