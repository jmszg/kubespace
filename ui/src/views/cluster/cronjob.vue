<template>
  <div>
    <clusterbar :titleName="titleName" :nsFunc="nsSearch" :nameFunc="nameSearch" :delFunc="delFunc"/>
    <div class="dashboard-container">
      <!-- <div class="dashboard-text"></div> -->
      <el-table
        ref="multipleTable"
        :data="cronjobs"
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
          min-width="65"
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
          min-width="50"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="schedule"
          label="定时"
          min-width="45"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="suspend"
          label="挂起"
          min-width="30"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="concurrency_policy"
          label="并发策略"
          min-width="35"
          show-overflow-tooltip>
        </el-table-column>
        <el-table-column
          prop="last_schedule_time"
          label="上一次执行"
          min-width="55"
          show-overflow-tooltip>
          <template slot-scope="scope">
            {{ scope.row.last_schedule_time }}
          </template>
        </el-table-column>
        <el-table-column
          prop="created"
          label="创建时间"
          min-width="47"
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
                <el-dropdown-item v-if="$editorRole()" @click.native.prevent="getCronJobYaml(scope.row.namespace, scope.row.name)">
                  <svg-icon style="width: 1.3em; height: 1.3em; line-height: 40px; vertical-align: -0.25em" icon-class="edit" />
                  <span style="margin-left: 5px;">修改</span>
                </el-dropdown-item>
                <el-dropdown-item v-if="$editorRole()" @click.native.prevent="deleteCronJobs([{namespace: scope.row.namespace, name: scope.row.name}])">
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
        <el-button plain @click="updateCronJob()" size="small">确 定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { Clusterbar } from '@/views/components'
import { ResType, listResource, watchResource, getResource, delResource, updateResource } from '@/api/cluster/resource'
import { Message } from 'element-ui'
import { Yaml } from '@/views/components'

export default {
  name: 'CronJob',
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
        titleName: ["CronJobs"],
        maxHeight: window.innerHeight - this.$contentHeight,
        loading: true,
        originCronJobs: [],
        search_ns: [],
        search_name: '',
        delFunc: undefined,
        delCronJobs: [],
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
  watch: {
    cluster: function() {
      this.fetchData()
    }
  },
  beforeDestroy() {
    if(this.clusterSSE) this.clusterSSE.disconnect()
  },
  computed: {
    cronjobs: function() {
      let dlist = []
      for (let p of this.originCronJobs) {
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
      this.originCronJobs = []
      const cluster = this.$store.state.cluster
      if (cluster) {
        listResource(cluster, ResType.CronJob).then(response => {
          this.loading = false
          this.originCronJobs = response.data ? response.data : []
          this.fetchSSE()
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
        this.clusterSSE = watchResource(this.$sse, this.cluster, ResType.CronJob, this.sseWatch, {process: true})
      }
    },
    sseWatch(newObj) {
      if (newObj) {
        let newUid = newObj.resource.uid
        let newRv = newObj.resource.resource_version
        if (newObj.event === 'add') {
          for(let i in this.originCronJobs) {
            let d = this.originCronJobs[i]
            if (d.uid === newUid) return
          }
          this.originCronJobs.push(newObj.resource)
        } else if (newObj.event === 'update') {
          for (let i in this.originCronJobs) {
            let d = this.originCronJobs[i]
            if (d.uid === newUid) {
              if (d.resource_version < newRv){
                let newDp = newObj.resource
                this.$set(this.originCronJobs, i, newDp)
              }
              break
            }
          }
        } else if (newObj.event === 'delete') {
          this.originCronJobs = this.originCronJobs.filter(( { uid } ) => uid !== newUid)
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
      this.$router.push({name: 'cronjobDetail', params: {namespace: namespace, cronjobName: name}})
    },
    getCronJobYaml: function(namespace, name) {
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
      getResource(cluster, ResType.CronJob, namespace, name, "yaml").then(response => {
        this.yamlLoading = false
        this.yamlValue = response.data
        this.yamlNamespace = namespace
        this.yamlName = name
      }).catch(() => {
        this.yamlLoading = false
      })
    },
    deleteCronJobs: function(cronjobs) {
      const cluster = this.$store.state.cluster
      if (!cluster) {
        Message.error("获取集群参数异常，请刷新重试")
        return
      }
      if ( cronjobs.length <= 0 ){
        Message.error("请选择要删除的CronJobs")
        return
      }
      let params = {
        resources: cronjobs
      }
      delResource(cluster, ResType.CronJob, params).then(() => {
        Message.success("删除成功")
      }).catch(() => {
        // console.log(e)
      })
    },
    updateCronJob: function() {
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
        Message.error("获取CronJob参数异常，请刷新重试")
        return
      }
      this.yamlLoading = true
      updateResource(cluster, ResType.CronJob, this.yamlNamespace, this.yamlName, this.yamlValue).then(() => {
        Message.success("更新成功")
        this.yamlLoading = false
        this.yamlDialog = false
      }).catch(() => {
        // console.log(e) 
      })
    },
    _delCronJobsFunc: function() {
      if (this.delCronJobs.length > 0){
        let delCronJobs = []
        for (var p of this.delCronJobs) {
          delCronJobs.push({namespace: p.namespace, name: p.name})
        }
        this.deleteCronJobs(delCronJobs)
      }
    },
    handleSelectionChange(val) {
      this.delCronJobs = val;
      if (val.length > 0){
        this.delFunc = this._delCronJobsFunc
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
