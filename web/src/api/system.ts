import { request } from '@/utils/request';

const Api = {
  // 系统状态
  GetStatus: '/api/v1/status',
};

// 系统状态响应
export interface SystemStatus {
  status: string;
  version: string;
  registration_enabled: boolean;
}

// 获取系统状态
export function getSystemStatus() {
  return request.get<SystemStatus>({
    url: Api.GetStatus,
  });
}
