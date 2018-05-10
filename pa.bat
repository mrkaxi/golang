@echo off 
set work_path=D:\wamp\www\discuz3.4\source\plugin\bbssdkpt
E: 
cd %work_path% 
for /R %%s in (.,*) do ( 
    dos2unix %%s 
)