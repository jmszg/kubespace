<template>
  <div v-loading="loading">
    <el-form :model="params" ref="job" label-position="left" label-width="105px">
      <el-form-item label="部署集群" prop="" :required="true">
        <el-select v-model="params.cluster" placeholder="请选择要部署的集群" size="small" style="width: 320px" @change="fetchNamespace">
          <el-option
            v-for="res in clusters"
            :key="res.name"
            :label="res.name1"
            :value="res.name">
          </el-option>
        </el-select>
      </el-form-item>
      <el-form-item label="命名空间" prop="" :required="true">
        <el-select v-model="params.namespace" placeholder="请选择要部署的命名空间" size="small" style="width: 320px">
          <el-option
            v-for="res in namespaces"
            :key="res.name"
            :label="res.name"
            :value="res.name">
          </el-option>
        </el-select>
      </el-form-item>
      <el-form-item label="Yaml" prop="">
        <div slot="label">
          Yaml
          <div style="margin-top: -15px;">
            (变量替换) 
            <el-popover placement="top-start" title="" width="500" trigger="hover">
              <div style="line-height: 24px;">
                  Yaml内容支持go template语法变量替换。比如：在上游或当前阶段配置了变量`a`，则在Yaml中可以使用{{ a }} (注意变量前要加.) 
                  替换为变量`a`的值。
              </div>
              <i slot="reference" class="el-icon-question"></i>
            </el-popover>
          </div>
        </div>
        <!-- <el-input type="textarea" :rows="17" v-model="params.yaml" placeholder="请输入Kubernetes Yaml内容" size="small"></el-input> -->
        <div>
          <yaml v-model="params.yaml" :loading="yamlLoading" :height="380"></yaml>
        </div>
      </el-form-item>
      
    </el-form>
  </div>
</template>

<script>
import { Yaml } from '@/views/components'
import { listCluster } from '@/api/cluster'
import { ResType, listResource } from '@/api/cluster/resource'

export default {
  name: 'DeployK8s',
  components: {
    Yaml,
  },
  data() {
    return {
      loading: false,
      resources: [],
      namespaces: [],
      clusters: [],
      yamlLoading: false,
      a: "{{ .a }}"
    }
  },
  props: ['params'],
  computed: {
  },
  beforeMount() {
    if(!this.params.namespace) {
      this.$set(this.params, 'namespace', 'default')
    }
    if(!this.params.yaml) {
      this.$set(this.params, 'yaml', "")
    }
    this.fetchClusters()
  },
  methods: {
    fetchClusters() {
      this.namespaces = []
      listCluster().then((response) => {
          this.clusters = response.data || [];
          if(this.params.cluster) this.fetchNamespace()
        }).catch(() => {
        })
    },
    fetchNamespace: function() {
      this.namespaces = []
      const cluster = this.params.cluster
      if (cluster) {
        listResource(cluster, ResType.Namespace).then(response => {
          this.namespaces = response.data
          this.namespaces.sort((a, b) => {return a.name > b.name ? 1 : -1})
        }).catch((err) => {
          console.log(err)
        })
      } else {
        Message.error("获取集群异常，请刷新重试")
      }
    },
  }
}
</script>

<style scoped>

</style>