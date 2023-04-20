<template>
  <div class="container-sm text-center">
    <h2 class="my-4">OpenAPI Viewer</h2>
    <form class="center">
      <div class="form-group row mb-3 justify-content-center">
        <label for="ui" class="form-label col-md-2 col-sm-2 col-xs-2">UI</label>
        <div class="col-md-3 col-sm-4 col-xs-5">
          <select id="ui" name="ui" v-model="ui" class="form-select">
            <option v-for="ui in uis" :value="ui">{{ ui }}</option>
          </select>
        </div>
      </div>
      <div class="form-group row mb-3 justify-content-center">
        <label for="apifile" class="form-label col-md-2 col-sm-2 col-xs-2">API File</label>
        <div class="col-md-3 col-sm-4 col-xs-5">
          <select id="apifile" name="apifile" v-model="apifile" class="form-select">
            <option v-for="file in apifiles" :value="file">{{ file }}</option>
          </select>
        </div>
      </div>
    </form>

    <div class="row mt-3 justify-content-center">
      <div class="col-md-3 col-sm-4 col-xs-5">
        <button type="button" @click="postView" class="btn btn-primary">View API</button>
      </div>

    </div>

  </div>
</template>

<script>
import axios from 'axios';

export default {
  data() {
    return {
      ui: "",
      apifile: "",
      uis: [],
      apifiles: []
    }
  },
  methods: {
    postView() {
      // axios send ajax Post /view --> 
      // server return status 302 and Location --> 
      // browser send ajax get {Location} automatically --> 
      // server return 200 -->
      // axios then func
      // so if we don't redirect in then func, then window redirect won't happen
      // thanks https://stackoverflow.com/questions/54500755/response-undefined-for-302-status-axios
      axios.post('http://localhost:8000/view', {
        ui: this.ui,
        apifile: this.apifile
      })
        .then(response => {
          if (response.status === 200) {
            window.location.href = response.request.responseURL;
          }
        })
        .catch(error => {
          console.log(error);
        });
    },
    getSupportedUIs() {
      axios.get('http://localhost:8000/uis')
        .then(response => {
          this.uis = response.data;
          if (this.uis.length > 0) {
            this.ui = this.uis[0];
          }
        })
        .catch(error => {
          console.log(error);
        });
    },
    getApiFiles() {
      axios.get('http://localhost:8000/apifiles')
        .then(response => {
          this.apifiles = response.data;
          if (this.apifiles.length > 0) {
            this.apifile = this.apifiles[0];
          }
        })
        .catch(error => {
          console.log(error);
        });
    }
  },
  mounted() {
    this.getApiFiles();
    this.getSupportedUIs();
  }
}
</script>