<template>
  <div>
    <clusterbar :titleName="titleName" :nsFunc="nsSearch" :nameFunc="nameSearch" :delFunc="delFunc"/>
    <div class="dashboard-container">
      <!-- <div class="dashboard-text"></div> -->
      <el-table
        ref="multipleTable"
        :data="jobs"
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
          min-width="60"
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
          min-width="40"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="ready_replicas"
          label="Pods"
          min-width="45"
          show-overflow-tooltip>
          <template slot-scope="scope">
            <span v-if="scope.row.active > 0" class="back-class">
              {{ scope.row.active }} Running
            </span>
            <span v-if="scope.row.succeeded > 0" class="back-class">
              {{ scope.row.succeeded }} Succeeded
            </span>
            <span v-if="scope.row.failed > 0" class="back-class">
              {{ scope.row.failed }} Failed
            </span>
          </template>
        </el-table-column>
        <el-table-column
          prop="completions"
          label="Completions"
          min-width="30"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="conditions"
          label="状态"
          min-width="25"
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
          min-width="35"
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
                <el-dropdown-item v-if="$editorRole()" @click.native.prevent="getJobYaml(scope.row.namespace, scope.row.name)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="edit" />
                  <span style="margin-left: 5px;">修改</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$editorRole()" @click.native.prevent="deleteJobs([{namespace: scope.row.namespace, name: scope.row.name}])">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="delete" />
                  <span style="margin-left: 5px;">删除</span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
          </template>
        </el-table-column>
      </el-table>
    </div>
    <el-dialog title="编辑" :visible.sync="yamlDialog" :close-on-click-modal="false" width="60%" top="55px">
      <yaml v-if="yamlDialog" v-model="yamlValue" :loading="yamlLoading"></yaml>
      <span slot="footer" class="dialog-footer">
        <el-button plain @click="yamlDialog = false" size="small">取 消</el-button>
        <el-button plain @click="updateJob()" size="small">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { sse } from '@/api/cluster'
import { ResType, listResource, watchResource, getResource, delResource, updateResource, patchResource } from '@/api/cluster/resource'
import { Message } from 'element-ui'
import { Yaml } from '@/views/components'

export default {
  name: 'Job',
  components: {
    Clusterbar,
    Yaml
  },
  data() {
      return {
        yamlDialog: false,
        yamlNamespace: "",
        yamlName: "",
        yamlValue: "",
        yamlLoading: true,
        cellStyle: {border: 0},
        titleName: ["Jobs"],
        maxHeight: window.innerHeight - this.$contentHeight,
        loading: true,
        originJobs: [],
        search_ns: [],
        search_name: '',
        delFunc: undefined,
        delJobs: [],
        clusterSSE: undefined
      }
  },
  created() {
    this.fetchData()
  },
  beforeDestroy() {
    if(this.clusterSSE) this.clusterSSE.disconnect()
  },
  watch: {
    cluster: function() {
      this.fetchData()
    }
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
  computed: {
    jobs: function() {
      let dlist = []
      for (let p of this.originJobs) {
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
      this.originJobs = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listResource(cluster, ResType.Job).then(response => {
          this.loading = false
          this.originJobs = response.data ? response.data : []
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
        this.clusterSSE = watchResource(this.$sse, this.cluster, ResType.Job, this.sseWatch, {process: true})
      }
    },
    sseWatch(newObj) {
      if (newObj) {
        let newUid = newObj.resource.uid
        let newRv = newObj.resource.resource_version
        if (newObj.event === 'add') {
          for(let i in this.originJobs) {
            let d = this.originJobs[i]
            if (d.uid === newUid) {
              return
            }
          }
          this.originJobs.push(newObj.resource)
        } else if (newObj.event === 'update') {
          for (let i in this.originJobs) {
            let d = this.originJobs[i]
            if (d.uid === newUid) {
              if (d.resource_version < newRv){
                let newDp = newObj.resource
                this.$set(this.originJobs, i, newDp)
              }
              break
            }
          }
        } else if (newObj.event === 'delete') {
          this.originJobs = this.originJobs.filter(( { uid } ) => uid !== newUid)
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
      this.$router.push({name: 'jobDetail', params: {namespace: namespace, jobName: name}})
    },
    getJobYaml: function(namespace, name) {
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
        Message.error("获取Deployment名称参数异常，请刷新重试")
        return
      }
      this.yamlValue = ""
      this.yamlDialog = true
      this.yamlLoading = true
      getResource(cluster, ResType.Job, namespace, name, "yaml").then(response => {
        this.yamlLoading = false
        this.yamlValue = response.data
        this.yamlNamespace = namespace
        this.yamlName = name
      }).catch(() => {
        this.yamlLoading = false
      })
    },
    deleteJobs: function(jobs) {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if ( jobs.length <= 0 ){
        Message.error("请选择要删除的Jobs")
        return
      }
      let params = {
        resources: jobs
      }
      delResource(cluster, ResType.Job, params).then(() => {
        Message.success("删除成功")
      }).catch(() => {
        // console.log(e)
      })
    },
    updateJob: function() {
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
        Message.error("获取Job参数异常，请刷新重试")
        return
      }
      this.yamlLoading = true
      updateResource(ResType.Job, cluster, ResType.Job, this.yamlNamespace, this.yamlName, this.yamlValue).then(() => {
        Message.success("更新成功")
        this.yamlLoading = false
        this.yamlDialog = false
      }).catch(() => {
        // console.log(e) 
      })
    },
    _delJobsFunc: function() {
      if (this.delJobs.length > 0){
        let delJobs = []
        for (var p of this.delJobs) {
          delJobs.push({namespace: p.namespace, name: p.name})
        }
        this.deleteJobs(delJobs)
      }
    },
    handleSelectionChange(val) {
      this.delJobs = val;
      if (val.length > 0){
        this.delFunc = this._delJobsFunc
      } else {
        this.delFunc = undefined
      }
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
