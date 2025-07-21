import { theme } from 'antd';

// 品牌色彩系统
export const colors = {
  primary: {
    50: '#f0f9ff',
    100: '#e0f2fe',
    200: '#bae6fd',
    300: '#7dd3fc',
    400: '#38bdf8',
    500: '#0ea5e9',
    600: '#0284c7',
    700: '#0369a1',
    800: '#075985',
    900: '#0c4a6e',
  },
  secondary: {
    50: '#faf5ff',
    100: '#f3e8ff',
    200: '#e9d5ff',
    300: '#d8b4fe',
    400: '#c084fc',
    500: '#a855f7',
    600: '#9333ea',
    700: '#7c3aed',
    800: '#6b21a8',
    900: '#581c87',
  },
  accent: {
    50: '#ecfeff',
    100: '#cffafe',
    200: '#a5f3fc',
    300: '#67e8f9',
    400: '#22d3ee',
    500: '#06b6d4',
    600: '#0891b2',
    700: '#0e7490',
    800: '#155e75',
    900: '#164e63',
  },
  success: {
    50: '#f0fdf4',
    100: '#dcfce7',
    200: '#bbf7d0',
    300: '#86efac',
    400: '#4ade80',
    500: '#22c55e',
    600: '#16a34a',
    700: '#15803d',
    800: '#166534',
    900: '#14532d',
  },
  warning: {
    50: '#fffbeb',
    100: '#fef3c7',
    200: '#fde68a',
    300: '#fcd34d',
    400: '#fbbf24',
    500: '#f59e0b',
    600: '#d97706',
    700: '#b45309',
    800: '#92400e',
    900: '#78350f',
  },
  error: {
    50: '#fef2f2',
    100: '#fee2e2',
    200: '#fecaca',
    300: '#fca5a5',
    400: '#f87171',
    500: '#ef4444',
    600: '#dc2626',
    700: '#b91c1c',
    800: '#991b1b',
    900: '#7f1d1d',
  },
  neutral: {
    50: '#f8fafc',
    100: '#f1f5f9',
    200: '#e2e8f0',
    300: '#cbd5e1',
    400: '#94a3b8',
    500: '#64748b',
    600: '#475569',
    700: '#334155',
    800: '#1e293b',
    900: '#0f172a',
  },
};

// Ant Design 主题配置
export const lightTheme = {
  algorithm: theme.defaultAlgorithm,
  token: {
    // 主色调
    colorPrimary: colors.primary[600],
    colorSuccess: colors.success[500],
    colorWarning: colors.warning[500],
    colorError: colors.error[500],
    colorInfo: colors.accent[500],
    
    // 背景色
    colorBgContainer: '#ffffff',
    colorBgElevated: '#ffffff',
    colorBgLayout: colors.neutral[50],
    colorBgSpotlight: colors.neutral[100],
    
    // 文字色
    colorText: colors.neutral[900],
    colorTextSecondary: colors.neutral[600],
    colorTextTertiary: colors.neutral[400],
    colorTextQuaternary: colors.neutral[300],
    
    // 边框色
    colorBorder: colors.neutral[200],
    colorBorderSecondary: colors.neutral[100],
    
    // 圆角
    borderRadius: 8,
    borderRadiusLG: 12,
    borderRadiusSM: 6,
    borderRadiusXS: 4,
    
    // 字体
    fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
    fontSize: 14,
    fontSizeLG: 16,
    fontSizeSM: 12,
    fontSizeXL: 20,
    
    // 间距
    padding: 16,
    paddingLG: 24,
    paddingSM: 12,
    paddingXS: 8,
    
    // 阴影
    boxShadow: '0 1px 2px 0 rgba(0, 0, 0, 0.03), 0 1px 6px -1px rgba(0, 0, 0, 0.02), 0 2px 4px 0 rgba(0, 0, 0, 0.02)',
    boxShadowSecondary: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
    boxShadowTertiary: '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
    
    // 动画
    motionDurationFast: '0.1s',
    motionDurationMid: '0.2s',
    motionDurationSlow: '0.3s',
  },
  components: {
    Layout: {
      headerBg: '#ffffff',
      headerHeight: 64,
      siderBg: '#ffffff',
      bodyBg: colors.neutral[50],
    },
    Menu: {
      itemBg: 'transparent',
      itemSelectedBg: colors.primary[50],
      itemSelectedColor: colors.primary[600],
      itemHoverBg: colors.neutral[100],
    },
    Button: {
      borderRadius: 8,
      controlHeight: 40,
      controlHeightLG: 48,
      controlHeightSM: 32,
    },
    Card: {
      borderRadius: 12,
      boxShadow: '0 1px 2px 0 rgba(0, 0, 0, 0.03), 0 1px 6px -1px rgba(0, 0, 0, 0.02)',
    },
    Input: {
      borderRadius: 8,
      controlHeight: 40,
      controlHeightLG: 48,
      controlHeightSM: 32,
    },
    Select: {
      borderRadius: 8,
      controlHeight: 40,
    },
    Table: {
      borderRadius: 8,
      headerBg: colors.neutral[50],
    },
    Modal: {
      borderRadius: 12,
    },
    Drawer: {
      borderRadius: 12,
    },
    Tabs: {
      borderRadius: 8,
    },
    Progress: {
      defaultColor: colors.primary[500],
    },
  },
};

export const darkTheme = {
  algorithm: theme.darkAlgorithm,
  token: {
    ...lightTheme.token,
    colorBgContainer: colors.neutral[800],
    colorBgElevated: colors.neutral[700],
    colorBgLayout: colors.neutral[900],
    colorBgSpotlight: colors.neutral[800],
    colorText: colors.neutral[100],
    colorTextSecondary: colors.neutral[300],
    colorTextTertiary: colors.neutral[400],
    colorBorder: colors.neutral[700],
    colorBorderSecondary: colors.neutral[800],
  },
  components: {
    ...lightTheme.components,
    Layout: {
      headerBg: colors.neutral[800],
      headerHeight: 64,
      siderBg: colors.neutral[800],
      bodyBg: colors.neutral[900],
    },
    Menu: {
      itemBg: 'transparent',
      itemSelectedBg: colors.primary[900],
      itemSelectedColor: colors.primary[400],
      itemHoverBg: colors.neutral[700],
    },
    Table: {
      borderRadius: 8,
      headerBg: colors.neutral[800],
    },
  },
};

// 响应式断点
export const breakpoints = {
  xs: '480px',
  sm: '576px',
  md: '768px',
  lg: '992px',
  xl: '1200px',
  xxl: '1600px',
};

// 动画配置
export const animations = {
  fadeIn: {
    initial: { opacity: 0 },
    animate: { opacity: 1 },
    exit: { opacity: 0 },
    transition: { duration: 0.2 },
  },
  slideUp: {
    initial: { opacity: 0, y: 20 },
    animate: { opacity: 1, y: 0 },
    exit: { opacity: 0, y: -20 },
    transition: { duration: 0.3 },
  },
  slideDown: {
    initial: { opacity: 0, y: -20 },
    animate: { opacity: 1, y: 0 },
    exit: { opacity: 0, y: 20 },
    transition: { duration: 0.3 },
  },
  scaleIn: {
    initial: { opacity: 0, scale: 0.9 },
    animate: { opacity: 1, scale: 1 },
    exit: { opacity: 0, scale: 0.9 },
    transition: { duration: 0.2 },
  },
};

// 工具函数
export const getTheme = (isDark: boolean) => isDark ? darkTheme : lightTheme;
