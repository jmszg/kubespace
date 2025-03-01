<template>
  <div>
    <clusterbar :titleName="titleName" :nameFunc="nameSearch" :createFunc="openCreateFormDialog" createDisplay="创建空间"/>
    <div class="dashboard-container" ref="tableCot">
      <el-table
        ref="multipleTable"
        :data="originWorkspaces"
        class="table-fix"
        :cell-style="cellStyle"
        v-loading="loading"
        :default-sort = "{prop: 'name'}"
        tooltip-effect="dark"
        style="width: 100%"
      >
        <el-table-column prop="name" label="名称" show-overflow-tooltip min-width="15">
          <template slot-scope="scope">
            <span class="name-class" v-on:click="nameClick(scope.row.id)">
              {{ scope.row.name }}
            </span>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" show-overflow-tooltip min-width="15">
        </el-table-column>
        <el-table-column prop="cluster_id" label="绑定集群" show-overflow-tooltip min-width="15">
          <template slot-scope="scope">
            {{ scope.row.cluster ? scope.row.cluster.name1 : scope.row.cluster_id }}
          </template>
        </el-table-column>
        <el-table-column prop="namespace" label="命名空间" show-overflow-tooltip min-width="15">
        </el-table-column>
        <el-table-column prop="owner" label="负责人" show-overflow-tooltip min-width="10">
          <template slot-scope="scope">
            {{ scope.row.owner }}
          </template>
        </el-table-column>
        <el-table-column prop="update_time" label="更新时间" show-overflow-tooltip min-width="15">
          <template slot-scope="scope">
            {{ $dateFormat(scope.row.update_time) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="170">
          <template slot-scope="scope">
            <div class="tableOperate">
              <el-link :underline="false" style="margin-right: 13px;" type="primary" @click="nameClick(scope.row.id)">详情</el-link>
              <el-link :disabled="!$editorRole(scope.row.id)" :underline="false" type="primary" style="margin-right: 13px" @click="openUpdateFormDialog(scope.row)">编辑</el-link>
              <el-link :disabled="!$adminRole(scope.row.id)" :underline="false" type="primary" style="margin-right: 13px" @click="openCloneFormDialog(scope.row)">克隆</el-link>
              <el-link :disabled="!$adminRole(scope.row.id)" :underline="false" type="danger" @click="openDeleteDialog(scope.row)">删除</el-link>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <el-dialog :title="updateFormVisible ? '修改空间' : form.clone ? '克隆空间-' + form.oriName : '创建空间'" :visible.sync="createFormVisible"
      @close="closeFormDialog" :destroy-on-close="true" :close-on-click-modal="false">
        <div v-loading="dialogLoading">
          <div class="dialogContent" style="">
            <el-alert v-if="form.clone" style="margin-bottom: 10px;" title="克隆会将工作空间中的应用以及K8s资源复制到新的工作空间" :closable="false" type="info"></el-alert>
            <el-form :model="form" :rules="rules" ref="form" label-position="left" label-width="105px">
              <el-form-item label="名称" prop="name" autofocus>
                <el-input v-model="form.name" autocomplete="off" placeholder="请输入空间名称" size="small"></el-input>
              </el-form-item>
              <el-form-item label="描述" prop="description">
                <el-input v-model="form.description" type="textarea" autocomplete="off" placeholder="请输入空间描述" size="small"></el-input>
              </el-form-item>
              <el-form-item label="集群" prop="" :required="true">
                <el-select v-model="form.cluster_id" placeholder="请选择要绑定的集群" size="small" style="width: 100%;"
                  @change="fetchNamespace" :disabled="updateFormVisible">
                  <el-option
                    v-for="item in clusters"
                    :key="item.name"
                    :label="item.name1"
                    :value="item.name"
                    :disabled="item.status != 'Connect'">
                  </el-option>
                </el-select>
              </el-form-item>
              <el-form-item label="命名空间" prop="" :required="true">
                <el-select v-model="form.namespace" placeholder="请选择要绑定的命名空间" size="small" style="width: 100%;"
                  :disabled="updateFormVisible">
                  <el-option
                    v-for="item in namespaces"
                    :key="item.name"
                    :label="item.name"
                    :value="item.name">
                  </el-option>
                </el-select>
              </el-form-item>
              <el-form-item label="负责人" prop="" :required="true">
                <el-input v-model="form.owner" autocomplete="off" placeholder="请输入该项目空间负责人" size="small"></el-input>
              </el-form-item>
            </el-form>
          </div>
          <div slot="footer" class="dialogFooter" style="margin-top: 20px;">
            <el-button @click="createFormVisible = false" style="margin-right: 20px;" >取 消</el-button>
            <el-button type="primary" @click="updateFormVisible ? handleUpdateWorkspace() : form.clone ? handleCloneWorkspace() : handleCreateWorkspace()" >确 定</el-button>
          </div>
        </div>
      </el-dialog>
      <el-dialog
        title="提示"
        :visible.sync="dialogDeleteVisible"
        width="30%" >
        <div v-loading="dialogLoading">
          <i class="el-icon-warning" style="font-size: 18px; margin-right: 10px;"></i>
          <span >请确认是否删除「{{ dialogDeleteObject.name }}」工作空间？</span>
          <el-checkbox style="margin: 15px 0px 0px 30px;" v-model="dialogDeleteObject.del_resource">同时删除集群资源</el-checkbox>
          <div slot="footer" class="dialog-footer" style="margin: 20px; text-align: right;">
            <el-button size="small" @click="dialogDeleteVisible = false">取 消</el-button>
            <el-button size="small" type="primary" @click="handleDeleteWorkspace">确 定</el-button>
          </div>
        </div>
      </el-dialog>
    </div>
  </div>
</template>
<script>
import { Clusterbar } from "@/views/components";
import { createProject, cloneProject, listProjects, updateProject, deleteProject } from "@/api/project/project";
import { listCluster } from '@/api/cluster'
import { ResType, listResource } from '@/api/cluster/resource'
import { Message } from "element-ui";

export default {
  name: "workspace",
  components: {
    Clusterbar,
  },
  mounted: function () {
    const that = this;
    window.onresize = () => {
      return (() => {
        let heightStyle = window.innerHeight - 135;
        that.maxHeight = heightStyle;
      })();
    };
  },
  data() {
    return {
      maxHeight: window.innerHeight - 135,
      cellStyle: { border: 0 },
      titleName: ["工作空间"],
      loading: true,
      dialogLoading: false,
      createFormVisible: false,
      updateFormVisible: false,
      clusters: [],
      namespaces: [],
      form: {
        clone: false,
        oriId: '',
        oriName: '',
        id: "",
        name: "",
        description: "",
        cluster_id: "",
        namespace: "",
        owner: this.$store.getters.username,
      },
      rules: {
        name: [{ required: true, message: '请输入空间名称', trigger: 'blur' },],
      },
      originWorkspaces: [],
      search_name: "",
      secretTypeMap: {
        'password': '密码',
        'key': '密钥',
        'token': 'AccessToken'
      },
      dialogDeleteVisible: false,
      dialogDeleteObject: {},
    };
  },
  created() {
    this.fetchWorkspaces();
  },
  computed: {
    secrets() {

    }
  },
  methods: {
    nameClick: function(id) {
      this.$router.push({name: 'workspaceOverview', params: {'workspaceId': id}})
    },
    handleEdit(index, row) {
      console.log(index, row);
    },
    fetchWorkspaces() {
      this.loading = true
      listProjects().then((resp) => {
        this.originWorkspaces = resp.data ? resp.data : []
        this.loading = false
      }).catch((err) => {
        console.log(err)
        this.loading = false
      })
    },
    handleCreateWorkspace() {
      if(!this.form.name) {
        Message.error("空间名称不能为空");
        return
      }
      if(!this.form.cluster_id) {
        Message.error("请选择要绑定的集群");
        return
      }
      if(!this.form.namespace) {
        Message.error("请选择要绑定的集群命名空间");
        return
      }
      if(!this.form.owner) {
        Message.error("空间负责人不能为空");
        return
      }
      let project = {
        name: this.form.name, 
        description: this.form.description, 
        cluster_id: this.form.cluster_id,
        namespace: this.form.namespace,
        owner: this.form.owner
      }
      this.dialogLoading = true
      createProject(project).then(() => {
        this.dialogLoading = false
        this.createFormVisible = false;
        Message.success("创建工作空间成功")
        this.fetchWorkspaces()
      }).catch((err) => {
        this.dialogLoading = false
        console.log(err)
      });
    },
    handleCloneWorkspace() {
      if(!this.form.name) {
        Message.error("空间名称不能为空");
        return
      }
      if(!this.form.cluster_id) {
        Message.error("请选择要绑定的集群");
        return
      }
      if(!this.form.namespace) {
        Message.error("请选择要绑定的集群命名空间");
        return
      }
      if(!this.form.owner) {
        Message.error("空间负责人不能为空");
        return
      }
      let project = {
        origin_project_id: this.form.oriId,
        name: this.form.name, 
        description: this.form.description, 
        cluster_id: this.form.cluster_id,
        namespace: this.form.namespace,
        owner: this.form.owner
      }
      this.dialogLoading = true
      cloneProject(project).then(() => {
        this.dialogLoading = false
        this.createFormVisible = false;
        Message.success("克隆工作空间成功")
        this.fetchWorkspaces()
      }).catch((err) => {
        this.dialogLoading = false
        this.createFormVisible = false;
        this.fetchWorkspaces()
      });
    },
    handleUpdateWorkspace() {
      if(!this.form.id) {
        Message.error("获取空间id参数异常，请刷新重试");
        return
      }
      let project = {
        name: this.form.name, 
        description: this.form.description, 
        owner: this.form.owner
      }
      if(!this.form.name) {
        Message.error("空间名称不能为空");
        return
      }
      if(!this.form.owner) {
        Message.error("空间负责人不能为空");
        return
      }
      this.dialogLoading = true
      updateProject(this.form.id, project).then(() => {
        this.dialogLoading = false
        this.createFormVisible = false;
        Message.success("更新工作空间成功")
        this.fetchWorkspaces()
      }).catch((err) => {
        this.dialogLoading = false
        console.log(err)
      });
    },
    openDeleteDialog(project) {
      this.dialogDeleteObject = {
        id: project.id,
        name: project.name,
        del_resource: true,
      }
      this.dialogDeleteVisible = true
    },
    handleDeleteWorkspace() {
      if(!this.dialogDeleteObject) {
        Message.error("获取密钥id参数异常，请刷新重试");
        return
      }
      this.dialogLoading = true
      deleteProject(this.dialogDeleteObject.id, {"del_resource": this.dialogDeleteObject.del_resource}).then(() => {
        this.dialogLoading = false
        Message.success("删除工作空间成功")
        this.dialogDeleteVisible = false
        this.fetchWorkspaces()
      }).catch((err) => {
        this.dialogLoading = false
        console.log(err)
      });
        
    },
    nameSearch(val) {
      this.search_name = val;
    },
    fetchClusters() {
      this.namespaces = []
      listCluster()
        .then((response) => {
          this.clusters = response.data || [];
        }).catch(() => {
        })
    },
    fetchNamespace: function() {
      this.namespaces = []
      const cluster = this.form.cluster_id
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
    openCreateFormDialog() {
      this.form = {
        name: "",
        description: "",
        cluster_id: "",
        namespace: "",
        owner: this.$store.getters.username,
      }
      this.createFormVisible = true;
      this.fetchClusters()
    },
    openUpdateFormDialog(project) {
      this.form = {
        id: project.id,
        name: project.name,
        description: project.description,
        cluster_id: project.cluster_id,
        namespace: project.namespace,
        owner: project.owner,
      }
      this.fetchClusters()
      this.updateFormVisible = true;
      this.createFormVisible = true;
    },
    openCloneFormDialog(project) {
      this.form = {
        clone: true,
        oriName: project.name,
        oriId: project.id,
        name: "",
        description: "",
        cluster_id: "",
        namespace: "",
        owner: this.$store.getters.username,
      }
      this.fetchClusters()
      this.createFormVisible = true;
    },
    closeFormDialog() {
      this.updateFormVisible = false; 
      this.createFormVisible = false;
    }
  },
};
</script>


<style lang="scss" scoped>
@import "~@/styles/variables.scss";

.table-fix {
  height: calc(100% - 100px);
}

</style>
