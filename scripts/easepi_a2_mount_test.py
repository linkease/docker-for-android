# coding:utf-8
from time import sleep

import time
import random
import os
import subprocess
from datetime import datetime

def get_current_version():
    output = subprocess.run("adb shell getprop ro.build.id", stdout=subprocess.PIPE).stdout.decode('utf-8')
    print("当前版本号：" + output)
    return output.rstrip("\n").rstrip("\r")

def check_device():
    # 查看adb设备存在
    buildid = subprocess.run("adb shell getprop ro.build.id", stdout=subprocess.PIPE).stdout.decode('utf-8')
    if "2025" in buildid:
        print("found adb device")
        isrecovery=subprocess.run("adb shell getprop init.svc.recovery", stdout=subprocess.PIPE).stdout.decode('utf-8')
        if "running" in isrecovery:
            print("FAIL: device is in recovery mode")
            return False
        else:
            return True
    else:
        print("FAIL: not found adb devices")
        return False

def check_device_avalible():
    # 查看adb设备存在
    check_num = 0
    device_status = True
    while not check_device():
        time.sleep(5)
        check_num += 1
        if check_num > 36:
            device_status = False
            print("设备连接超时")
            break

    if not device_status:
        return False
    else:
        return True

def random_delay(num):
    # 获取30到300之间的随机整数
    random_number = num + 5
    if random_number > 360:
        random_number = 360
    print(f"生成的随机数是: {random_number}s")
    time.sleep(random_number)

if __name__ == "__main__":

    # 循环测试
    i = 1
    try:
        while i < 100:
            print(f"当前测试次数: {i}")
            time.sleep(3)
            # 检查是否识别到设备
            if not check_device_avalible():
                break

            os.system("adb root")

            # 检查android系统是否启动，如果没有启动完成，循环等待
            
            android_status = subprocess.run("adb shell getprop sys.boot_completed", stdout=subprocess.PIPE).stdout.decode('utf-8')
            while "1" not in android_status:
                print("android not booted, waiting...")
                time.sleep(5)
                android_status = subprocess.run("adb shell getprop sys.boot_completed", stdout=subprocess.PIPE).stdout.decode('utf-8')
            print("PASS: android booted")

            # 延迟60秒
            time.sleep(60)

            # 检查dockerd是否启动
            dockerd_status = subprocess.run("adb shell ps -ef | grep dockerd", stdout=subprocess.PIPE).stdout.decode('utf-8')
            if "dockerd" not in dockerd_status:
                print("FAIL: dockerd not running")
            else:
                print("PASS: dockerd running")

            # adb pull /data/local/docker/start-dockerd.log 到当前目录，并按照时间命名
            current_time = datetime.now().strftime("%Y%m%d_%H%M%S")
            os.system(f"adb pull /data/local/docker/start-dockerd.log ./start-dockerd_{current_time}.log")

            # 随机延迟
            # random_delay(i+60)
            time.sleep(60)
            # 重启设备
            os.system("adb reboot")
            time.sleep(3)
            i = i + 1
    except KeyboardInterrupt:
        print('程序被用户中断')
