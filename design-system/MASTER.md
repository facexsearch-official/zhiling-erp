# PISA 进销存 - 设计系统规范

> 基于 ui-ux-pro-max 规则，为 B2B SaaS 进销存系统定制

---

## 一、风格定位

### 推荐风格：**Minimalism + Flat Design（极简扁平）**

| 维度 | 选择 | 理由 |
|---|---|---|
| 风格 | Minimalism（极简） | 数据密集型工具，减少视觉噪音，提升信息密度 |
| 辅助 | Flat Design（扁平） | 专业感强，渲染性能好，适合 B2B SaaS |
| 不选 | Glassmorphism / Neumorphism / Brutalism | 装饰性强，干扰数据阅读，不适合工具类产品 |

**核心原则**：
- 数据优先，装饰最小化
- 高信息密度，低视觉噪音
- 一致性高于创意
- 效率优于美观

---

## 二、色彩系统

### 2.1 主色板（Primary）

| Token | 色值 | 用途 |
|---|---|---|
| `primary-50` | `#EFF6FF` | 浅底色/悬浮行 |
| `primary-100` | `#DBEAFE` | 选中行背景 |
| `primary-200` | `#BFDBFE` | 边框高亮 |
| `primary-300` | `#93C5FD` | 禁用态 |
| `primary-400` | `#60A5FA` | 图标悬停 |
| `primary-500` | `#3B82F6` | 主按钮/链接/活跃态 |
| `primary-600` | `#2563EB` | 主按钮悬停 |
| `primary-700` | `#1D4ED8` | 主按钮按下 |
| `primary-800` | `#1E40AF` | 标题强调 |
| `primary-900` | `#1E3A8A` | 深色文字 |

**语义**：蓝色 = 信任/专业/稳定，B2B 产品标配

### 2.2 语义色（Semantic）

| Token | 色值 | 用途 |
|---|---|---|
| `success-500` | `#22C55E` | 成功/入库/盈利 |
| `success-50` | `#F0FDF4` | 成功背景 |
| `warning-500` | `#F59E0B` | 警告/库存预警 |
| `warning-50` | `#FFFBEB` | 警告背景 |
| `danger-500` | `#EF4444` | 错误/删除/亏损 |
| `danger-50` | `#FEF2F2` | 错误背景 |
| `info-500` | `#06B6D4` | 信息提示 |
| `info-50` | `#ECFEFF` | 信息背景 |

### 2.3 中性色（Neutral）

| Token | 色值 | 用途 |
|---|---|---|
| `gray-50` | `#F9FAFB` | 页面背景 |
| `gray-100` | `#F3F4F6` | 卡片背景/侧边栏 |
| `gray-200` | `#E5E7EB` | 边框/分割线 |
| `gray-300` | `#D1D5DB` | 禁用边框 |
| `gray-400` | `#9CA3AF` | 占位符文字 |
| `gray-500` | `#6B7280` | 次要文字 |
| `gray-600` | `#4B5563` | 正文文字 |
| `gray-700` | `#374151` | 标题文字 |
| `gray-800` | `#1F2937` | 强调文字 |
| `gray-900` | `#111827` | 标题/重要数字 |

### 2.4 功能色映射

| 业务场景 | 颜色 | 说明 |
|---|---|---|
| 销售出货 | `primary-500` | 正常出库 |
| 销售退货 | `danger-500` | 红色警示 |
| 进货入库 | `success-500` | 绿色入库 |
| 采购退货 | `warning-500` | 黄色警告 |
| 库存预警 | `danger-500` | 低于最低库存 |
| 应收欠款 | `warning-500` | 待收款项 |
| 应付欠款 | `danger-500` | 待付款项 |
| 盈利 | `success-500` | 正利润 |
| 亏损 | `danger-500` | 负利润 |

---

## 三、字体系统

### 3.1 字体族

| 场景 | 字体 | 回退 |
|---|---|---|
| 中文正文 | `Noto Sans SC` | `PingFang SC`, `Microsoft YaHei`, `sans-serif` |
| 英文/数字 | `Inter` | `system-ui`, `sans-serif` |
| 代码/编号 | `JetBrains Mono` | `Fira Code`, `monospace` |

### 3.2 字号系统

| Token | 大小 | 行高 | 用途 |
|---|---|---|---|
| `text-2xs` | 10px | 14px | 角标/辅助标签 |
| `text-xs` | 12px | 16px | 辅助文字/表格次要信息 |
| `text-sm` | 14px | 20px | 表格正文/表单标签 |
| `text-base` | 16px | 24px | 页面正文 |
| `text-lg` | 18px | 28px | 小标题 |
| `text-xl` | 20px | 28px | 卡片标题 |
| `text-2xl` | 24px | 32px | 页面标题 |
| `text-3xl` | 30px | 36px | 数字统计 |

### 3.3 字重

| Token | 值 | 用途 |
|---|---|---|
| `font-normal` | 400 | 正文 |
| `font-medium` | 500 | 标签/按钮 |
| `font-semibold` | 600 | 小标题/表头 |
| `font-bold` | 700 | 页面标题/重要数字 |

---

## 四、间距系统（8px 基数）

| Token | 值 | 用途 |
|---|---|---|
| `space-0` | 0 | — |
| `space-1` | 4px | 图标与文字间距 |
| `space-2` | 8px | 紧凑间距/表格单元格 |
| `space-3` | 12px | 表单字段间距 |
| `space-4` | 16px | 卡片内边距/列表项间距 |
| `space-5` | 20px | 区块间距 |
| `space-6` | 24px | 卡片间距 |
| `space-8` | 32px | 页面区块间距 |
| `space-10` | 40px | 页面顶部间距 |
| `space-12` | 48px | 大区块间距 |

---

## 五、圆角系统

| Token | 值 | 用途 |
|---|---|---|
| `radius-none` | 0 | — |
| `radius-sm` | 4px | 小按钮/标签/输入框 |
| `radius-md` | 6px | 卡片/弹窗/下拉菜单 |
| `radius-lg` | 8px | 大卡片/面板 |
| `radius-xl` | 12px | 模态框 |
| `radius-full` | 9999px | 圆形头像/徽章 |

---

## 六、阴影系统

| Token | 值 | 用途 |
|---|---|---|
| `shadow-xs` | `0 1px 2px rgba(0,0,0,0.05)` | 按钮/输入框 |
| `shadow-sm` | `0 1px 3px rgba(0,0,0,0.1), 0 1px 2px rgba(0,0,0,0.06)` | 卡片/下拉 |
| `shadow-md` | `0 4px 6px rgba(0,0,0,0.1), 0 2px 4px rgba(0,0,0,0.06)` | 弹窗/浮层 |
| `shadow-lg` | `0 10px 15px rgba(0,0,0,0.1), 0 4px 6px rgba(0,0,0,0.05)` | 模态框 |
| `shadow-xl` | `0 20px 25px rgba(0,0,0,0.1), 0 10px 10px rgba(0,0,0,0.04)` | 浮动面板 |

---

## 七、组件规范

### 7.1 按钮

| 类型 | 样式 | 用途 |
|---|---|---|
| Primary | bg: primary-500, text: white, radius: sm | 主要操作（新增、保存、提交） |
| Secondary | bg: white, border: gray-200, text: gray-700 | 次要操作（取消、返回） |
| Danger | bg: danger-500, text: white | 危险操作（删除、作废） |
| Ghost | bg: transparent, text: gray-600 | 辅助操作（更多、筛选） |
| Text | bg: transparent, text: primary-500, underline | 链接操作 |

**按钮尺寸**：
| 尺寸 | 高度 | 内边距 | 字号 |
|---|---|---|---|
| sm | 32px | 8px 12px | 12px |
| md | 36px | 8px 16px | 14px |
| lg | 40px | 10px 20px | 14px |

### 7.2 表格

| 属性 | 值 |
|---|---|
| 表头背景 | `gray-50` |
| 表头字重 | `font-semibold` |
| 行高 | 48px |
| 悬浮行背景 | `primary-50` |
| 选中行背景 | `primary-100` |
| 边框 | `1px solid gray-200` |
| 单元格内边距 | 12px 16px |
| 数字对齐 | 右对齐（tabular-nums） |
| 文字对齐 | 左对齐 |

### 7.3 表单

| 属性 | 值 |
|---|---|
| 输入框高度 | 36px |
| 输入框边框 | `1px solid gray-300` |
| 聚焦边框 | `1px solid primary-500` + `ring: 3px primary-100` |
| 错误边框 | `1px solid danger-500` |
| 标签位置 | 输入框上方 |
| 标签字号 | 14px, font-medium |
| 错误提示 | 输入框下方, danger-500, 12px |
| 帮助文字 | 输入框下方, gray-500, 12px |

### 7.4 卡片

| 属性 | 值 |
|---|---|
| 背景 | white |
| 边框 | `1px solid gray-200` |
| 圆角 | 6px |
| 阴影 | shadow-sm |
| 内边距 | 16px 或 24px |
| 标题 | 16px, font-semibold, gray-800 |

### 7.5 导航侧边栏

| 属性 | 值 |
|---|---|
| 宽度 | 240px（展开）/ 64px（收起） |
| 背景 | white |
| 边框 | 右侧 `1px solid gray-200` |
| 菜单项高度 | 40px |
| 菜单项内边距 | 8px 12px |
| 活跃项背景 | `primary-50` |
| 活跃项文字 | `primary-600` |
| 活跃项左侧指示 | 3px solid primary-500 |
| 悬浮项背景 | `gray-50` |
| 图标大小 | 20px |
| 图标与文字间距 | 8px |

### 7.6 顶部栏

| 属性 | 值 |
|---|---|
| 高度 | 56px |
| 背景 | white |
| 底部边框 | `1px solid gray-200` |
| 内边距 | 0 24px |
| 标题字号 | 18px, font-semibold |

### 7.7 弹窗/模态框

| 属性 | 值 |
|---|---|
| 遮罩层 | `rgba(0,0,0,0.5)` |
| 弹窗背景 | white |
| 圆角 | 8px |
| 阴影 | shadow-xl |
| 最小宽度 | 400px |
| 最大宽度 | 640px |
| 内边距 | 24px |
| 标题字号 | 18px, font-semibold |
| 按钮间距 | 12px |

### 7.8 标签/徽章

| 类型 | 样式 |
|---|---|
| 默认 | bg: gray-100, text: gray-700 |
| 成功 | bg: success-50, text: success-500 |
| 警告 | bg: warning-50, text: warning-500 |
| 危险 | bg: danger-50, text: danger-500 |
| 信息 | bg: info-50, text: info-500 |
| 圆角 | radius-sm (4px) |
| 内边距 | 2px 8px |
| 字号 | 12px |

---

## 八、图标规范

| 属性 | 值 |
|---|---|
| 图标库 | Lucide Icons（推荐）/ Heroicons |
| 图标尺寸 | 16px（小）、20px（默认）、24px（大） |
| 线宽 | 1.5px |
| 颜色 | 继承父元素 color |
| 与文字间距 | 8px |

**禁止使用 Emoji 作为功能图标**

---

## 九、布局规范

### 9.1 整体布局

```
┌──────────────────────────────────────────────────────┐
│  顶部栏 (56px)                                        │
├──────────┬───────────────────────────────────────────┤
│ 侧边栏   │  内容区域                                  │
│ (240px)  │  ┌─────────────────────────────────────┐  │
│          │  │  页面标题 + 操作按钮                    │  │
│          │  ├─────────────────────────────────────┤  │
│          │  │  统计卡片 / 筛选栏                      │  │
│          │  ├─────────────────────────────────────┤  │
│          │  │  数据表格 / 表单内容                    │  │
│          │  ├─────────────────────────────────────┤  │
│          │  │  分页器                               │  │
│          │  └─────────────────────────────────────┘  │
└──────────┴───────────────────────────────────────────┘
```

### 9.2 响应式断点

| 断点 | 宽度 | 布局 |
|---|---|---|
| mobile | < 768px | 侧边栏隐藏，底部导航 |
| tablet | 768px - 1024px | 侧边栏收起（64px） |
| desktop | > 1024px | 侧边栏展开（240px） |

### 9.3 页面最大宽度

| 内容类型 | 最大宽度 |
|---|---|
| 表格页面 | 无限制（全宽） |
| 表单页面 | 720px 居中 |
| 详情页面 | 960px 居中 |
| 设置页面 | 640px 居中 |

---

## 十、暗色模式（Dark Mode）

| 亮色 | 暗色 |
|---|---|
| `white` | `#111827` (gray-900) |
| `gray-50` | `#1F2937` (gray-800) |
| `gray-100` | `#374151` (gray-700) |
| `gray-200` | `#4B5563` (gray-600) |
| `gray-600` | `#D1D5DB` (gray-300) |
| `gray-800` | `#F3F4F6` (gray-100) |
| `primary-500` | `#60A5FA` (primary-400) |
| `success-500` | `#4ADE80` (success-400) |
| `danger-500` | `#F87171` (danger-400) |

**暗色模式规则**：
- 文字对比度 ≥ 4.5:1
- 不使用纯白文字
- 表格行用灰度区分，不用色彩
- 边框用 `gray-700` 替代 `gray-200`

---

## 十一、动画规范

| 属性 | 值 |
|---|---|
| 持续时间 | 150ms - 300ms |
| 缓动函数 | `ease-out`（进入）/ `ease-in`（退出） |
| 悬浮过渡 | `transition: background-color 150ms ease` |
| 弹窗进入 | `fade-in + scale(0.95→1)` |
| 侧边栏收起 | `width transition 200ms ease` |
| 表格排序 | 无动画（即时切换） |
| 减弱动效 | 尊重 `prefers-reduced-motion` |

---

## 十二、数据可视化

### 推荐图表类型

| 场景 | 图表类型 | 库 |
|---|---|---|
| 销售趋势 | 折线图 | Chart.js |
| 销售额对比 | 柱状图 | Chart.js |
| 库存分布 | 柱状图（水平） | Chart.js |
| 成本占比 | 环形图 | Chart.js |
| 收付款趋势 | 面积图 | Chart.js |
| 营业员业绩 | 柱状图 | Chart.js |

### 图表配色

```javascript
const chartColors = {
  primary: '#3B82F6',
  success: '#22C55E',
  warning: '#F59E0B',
  danger: '#EF4444',
  info: '#06B6D4',
  purple: '#8B5CF6',
  pink: '#EC4899',
  indigo: '#6366F1',
};
```

### 图表规则

- 显示图例（Legend）
- 悬浮显示工具提示（Tooltip）
- 坐标轴标签带单位
- 空数据显示"暂无数据"
- 加载时显示骨架屏
- 数字使用千分位格式

---

## 十三、反模式（禁止）

| ❌ 禁止 | ✅ 替代 |
|---|---|
| Emoji 作为功能图标 | 使用 Lucide SVG 图标 |
| 纯色块传达信息 | 色彩 + 图标/文字双重提示 |
| 超过 5 种主色 | 使用语义色系统 |
| 表格内容居中 | 文字左对齐，数字右对齐 |
| placeholder 作为标签 | 使用 visible label |
| 悬浮才显示操作 | 始终显示主要操作按钮 |
| 无限滚动表格 | 分页器 |
| 弹窗内嵌弹窗 | 单层模态框 |
| 动画装饰 | 动画表达因果关系 |
| 禁用缩放 | 支持 viewport 缩放 |

---

## 十四、技术实现建议

### Flutter 实现

```dart
// 主题配置示例
class AppTheme {
  static const primaryColor = Color(0xFF3B82F6);
  static const successColor = Color(0xFF22C55E);
  static const warningColor = Color(0xFFF59E0B);
  static const dangerColor = Color(0xFFEF4444);

  static const gray50 = Color(0xFFF9FAFB);
  static const gray100 = Color(0xFFF3F4F6);
  static const gray200 = Color(0xFFE5E7EB);
  static const gray300 = Color(0xFFD1D5DB);
  static const gray500 = Color(0xFF6B7280);
  static const gray600 = Color(0xFF4B5563);
  static const gray700 = Color(0xFF374151);
  static const gray800 = Color(0xFF1F2937);
  static const gray900 = Color(0xFF111827);

  static const fontSize12 = 12.0;
  static const fontSize14 = 14.0;
  static const fontSize16 = 16.0;
  static const fontSize18 = 18.0;
  static const fontSize20 = 20.0;
  static const fontSize24 = 24.0;
  static const fontSize30 = 30.0;
}
```

### Tailwind CSS 对应（Web）

```css
/* 自定义 CSS 变量 */
:root {
  --primary-50: #EFF6FF;
  --primary-500: #3B82F6;
  --primary-600: #2563EB;
  --gray-50: #F9FAFB;
  --gray-100: #F3F4F6;
  --gray-200: #E5E7EB;
  --gray-600: #4B5563;
  --gray-800: #1F2937;
}
```

---

*设计系统生成于 2026-09-12*
*基于 ui-ux-pro-max 规则，适配 B2B SaaS 进销存系统*
