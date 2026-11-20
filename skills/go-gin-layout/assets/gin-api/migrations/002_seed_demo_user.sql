-- 仅用于本地示例：预置一个可供订单接口使用的用户。
-- 生产环境请改用注册流程或正式的初始化数据，不要执行本文件。
INSERT INTO t_user (id, name) VALUES (1, 'Alice')
ON DUPLICATE KEY UPDATE name = 'Alice';
