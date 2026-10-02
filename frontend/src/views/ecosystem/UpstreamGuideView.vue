<template>
  <div class="page">
    <div class="page-header">
      <h2 class="page-title">生态互联使用介绍</h2>
    </div>
    <el-card>
      <el-tabs v-model="tab">
        <el-tab-pane label="功能介绍" name="intro">
          <div class="doc">
            <h3>什么是生态互联</h3>
            <p>生态互联帮助你与上游供应商企业打通数据：连接上游企业后，供应商推送的商品可以一键接收为本地商品档案，商城装修方案也可以直接接收并应用到你的商城。</p>
            <h3>包含的能力</h3>
            <ul>
              <li><b>供应商连接</b>：与上游企业建立连接，并把连接关联到本地供应商档案。</li>
              <li><b>商品接收</b>：接收上游推送的商品资料（编号、条码、名称、规格、品牌、图片），确认接收后自动创建商品档案。</li>
              <li><b>商城装修接收</b>：接收上游或平台推送的商城装修方案，一键应用到商城。</li>
            </ul>
          </div>
        </el-tab-pane>
        <el-tab-pane label="如何使用" name="howto">
          <div class="doc">
            <h3>第一步：建立供应商连接</h3>
            <p>进入「供应商连接」，新增连接并填写上游企业微信名称与企业编号，然后选择要关联的本地供应商，点击连接。</p>
            <h3>第二步：接收商品</h3>
            <p>进入「商品接收」，在「自动接收」标签页查看上游推送的商品，确认无误后点击「接收」，系统会自动创建商品档案（含分类「供应商接收」与对应品牌）。不需要的商品可点击「关闭」。</p>
            <h3>第三步：接收商城装修</h3>
            <p>进入「商城装修接收」，逐条接收或点击「一键接收」将全部待接收装修方案应用到商城。</p>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const tab = ref('intro')
</script>

<style scoped>
.page { display: flex; flex-direction: column; gap: 12px; }
.page-header { display: flex; justify-content: space-between; align-items: center; }
.page-title { font-size: 18px; font-weight: 600; margin: 0; }
.doc { max-width: 760px; line-height: 1.8; color: #303133; }
.doc h3 { margin: 18px 0 8px; font-size: 15px; }
.doc ul { padding-left: 20px; }
</style>
