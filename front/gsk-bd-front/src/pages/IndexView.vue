<template>
  <q-page class="row justify-center q-pa-md">
    <div class="col-12 col-md-6 q-gutter-md">
      <q-select
        filled
        v-model="selectedBatFile"
        :options="batchFiles"
        label="选择 BAT 文件"
        @update:model-value="parseSelectedBatFile"
        class="q-mb-md"
      />
      <q-input filled v-model="params.a" label="HTTP API 地址 (-a)" />
      <q-select filled v-model="params.p" :options="textFiles" label="群列表文件名 (-p)" />
      <q-select filled v-model="params.w" :options="textFiles" label="要发送的信息 (-w)" />
      <q-input filled type="number" v-model="params.d" label="信息推送时间间隔 (-d)" />
      <q-input filled type="number" v-model="params.c" label="每个群推送的概率 (-c)" />
      <q-toggle filled v-model="params.h" label="显示帮助信息 (-h)" />
      <q-select filled v-model="params.s" :options="textFiles" label="保存文件路径 (-s)" />
      <q-toggle filled v-model="params.g" label="Gensokyo子频道过滤 (-g)" />
      <q-toggle filled v-model="params.f" label="私聊模式 (-f)" />
      <q-input filled v-model="params.t" label="Access Token (-t)" />
      <q-input filled v-model="params.b" label="本次存档名(输入后点创建存档)然后在保存文件路径 (-s)选中" />
      <q-toggle filled v-model="params.r" label="打乱列表顺序 (-r)" />
      <q-btn label="发送请求" color="primary" @click="sendRequest" />
      <q-btn label="创建存档" @click="createSave" color="primary" class="q-mt-md" />
    </div>
  </q-page>
</template>


<script setup>
import { ref } from 'vue';
import axios from 'axios';

const textFiles = ref([]);
const batchFiles = ref([]);

const params = ref({
  a: '',
  p: '',
  w: '',
  d: 10,
  c: 100,
  h: false,
  s: '',
  g: false,
  f: false,
  t: '',
  r: false,
  b: '',
});

async function sendRequest() {
 // 处理参数，确保所有值都是字符串
 function processParams(params) {
    const processed = {};
    for (const key in params) {
      const value = params[key];
      // 检查值是否为对象并且含有 value 属性
      if (value !== null && typeof value === 'object' && 'value' in value) {
        processed[key] = value.value;
      } else {
        processed[key] = value;
      }
    }
    return processed;
  }

  // 在发送请求前处理 params，移除 .txt 后缀
  const cleanedParams = processParams(params.value);
  const processedParams = {
    ...cleanedParams,
    p: cleanedParams.p.replace(/\.txt$/, ''),
    s: cleanedParams.s.replace(/\-save.txt$/, '').replace(/\.txt$/, '')
  };

  const queryString = Object.keys(processedParams)
    .filter(key => processedParams[key] !== false && processedParams[key] !== '')
    .map(key => `${key}=${encodeURIComponent(processedParams[key])}`)
    .join('&');
  const url = `/webui/api/run?${queryString}`;

  try {
    const response = await axios.get(url, {
      withCredentials: true // 确保请求与当前域关联，携带cookie
    });
    console.log('Request was successful', response.data);
  } catch (error) {
    console.error('Error during the API request:', error);
  }
}

async function loadFileList() {
  try {
    const response = await axios.get('/webui/api/list-files');
    // 添加一个空白选项至 textFiles 数组的开头
    const emptyOption = { label: 'None', value: '' }; // 自定义空白选项的显示文本和值

    // 确保这里使用的是小写的 'filename' 和 'content'
    textFiles.value = [emptyOption, ...response.data.textFiles.map(file => ({ label: file.filename, value: file.filename }))];
    batchFiles.value = response.data.batchFiles.map(file => ({ label: file.filename, value: file.content }));

    // 调试输出，查看加载的数据结构
    console.log('Loaded text files:', textFiles.value);
    console.log('Loaded batch files:', batchFiles.value);
  } catch (error) {
    console.error('Error loading file lists:', error);
  }
}

function parseSelectedBatFile(selectedObject) {
  // 首先检查selectedObject是否存在并且是一个对象
  if (selectedObject && typeof selectedObject === 'object') {
    console.log('Parsing BAT file content:', selectedObject.value);
    // 再次确认value是字符串
    if (typeof selectedObject.value === 'string') {
      parseContent(selectedObject.value);
    } else {
      console.error('Error: value property is not a string:', selectedObject.value);
    }
  } else {
    console.error('Error: Expected object, got:', selectedObject);
  }
}

function parseContent(content) {
  console.log('Parsing content:', content);
  const regex = /-[a-z]\s(?:"([^"]*)"|(\S+))/g;
  let match;

  while ((match = regex.exec(content)) !== null) {
    const paramKey = match[0][1]; // '-'后的第一个字符是参数键
    const paramValue = match[1] || match[2]; // 第一个捕获组是引号内的内容，第二个是非空格的内容
    if (paramKey && paramValue) {
      params.value[paramKey] = paramValue.replace(/^"|"$/g, ''); // 移除可能的引号
      console.log(`Param ${paramKey}: ${paramValue}`);
    }
  }
}

// Method to call the API for creating a new save
const createSave = async () => {
  if (!params.value.b) {
    console.error('存档文件名不能为空');
    return; // Early return if the filename is empty
  }
  try {
    const response = await axios.post('/webui/api/new-save', { filename: params.value.b });
    console.log('存档创建成功:', response.data);
    // 将新创建的文档名加入到textFiles中
    const newFileName = `${params.value.b}-save.txt`;
    textFiles.value.push({ label: newFileName, value: newFileName });
    console.log('更新后的文件列表:', textFiles.value);
    // 清空输入框
    params.value.b = '';
  } catch (error) {
    console.error('存档创建失败:', error);
  }
};

loadFileList();
</script>

<style scoped>
</style>
