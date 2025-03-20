这是一个命令行个人笔记本, 可以在 github/gitee 中创建自己的私人仓库作为云笔记存储


1. 本地运行所需go >= 1.23.5
   
2. 修改cfg/note_config.yaml 文件
   
  app:  
  
    db: "/Users/mo/WorkStation/go/note/db 修改为自己本地地址"  
    
    editor: "vim"  
    

  github:  
  
    url: "github/gitee 修改为私人仓库地址"  
    
    user: "username"  
    
    password: "token/password"  
    
    branch: "main"


3. 执行 sudo ./install

4. 执行 note 即可显示使用方法

5. 效果图

   <img width="950" alt="image" src="https://github.com/user-attachments/assets/df44d456-9570-4a20-b937-f5bc9b360edb" />

   <img width="578" alt="image" src="https://github.com/user-attachments/assets/2fe1ea5f-4031-4dc2-86f0-bb6c2117bdfd" />
   <img width="651" alt="image" src="https://github.com/user-attachments/assets/73377457-39cd-4181-9c0d-10baca863028" />


