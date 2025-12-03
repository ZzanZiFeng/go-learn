package main

import "fmt"

// // 1. 定义一个复杂类型：User 结构体
// type User struct {
// 	ID       int
// 	Name     string
// 	Email    string
// 	IsActive bool
// }

// func main() {
// 	// 2. 声明一个 User 结构体的切片

// 	// 方式 A: 声明一个 nil 切片 (推荐，append 时会自动分配)
// 	var users []User
// 	fmt.Printf("方式 A: users = %v, len = %d, cap = %d (是否为 nil: %t)\n", users, len(users), cap(users), users == nil)

// 	// 方式 B: 使用 make() 声明一个空切片 (长度为 0，但容量可能不为 0)
// 	activeUsers := make([]User, 0, 5) // 预分配容量为 5，可以提升性能
// 	fmt.Printf("方式 B: activeUsers = %v, len = %d, cap = %d (是否为 nil: %t)\n", activeUsers, len(activeUsers), cap(activeUsers), activeUsers == nil)

// 	// 方式 C: 使用字面量声明并初始化
// 	initialUsers := []User{
// 		{ID: 1, Name: "Alice", Email: "alice@example.com", IsActive: true},
// 		{ID: 2, Name: "Bob", Email: "bob@example.com", IsActive: false},
// 	}
// 	fmt.Printf("方式 C: initialUsers = %v, len = %d, cap = %d\n", initialUsers, len(initialUsers), cap(initialUsers))

// 	// 向 nil 切片中添加元素
// 	user1 := User{ID: 3, Name: "Charlie", Email: "charlie@example.com", IsActive: true}
// 	users = append(users, user1)

// 	fmt.Println("\n添加一个用户后:")
// 	fmt.Println("users:", users)

// 	// 遍历切片
// 	fmt.Println("\n遍历 initialUsers:")
// 	for _, u := range initialUsers {
// 		if u.IsActive {
// 			fmt.Printf("- Active User: %s (%s)\n", u.Name, u.Email)
// 		}
// 	}
// }

// 新的复杂多层级数组声明示例
// 这个示例展示了如何在Go语言中创建和操作复杂的多层级数据结构
func main() {
	// 1. 定义多个结构体类型，用于创建多层级数据结构
	// 在Go中，type关键字用于定义新的类型，struct用于定义结构体类型
	// 结构体是Go中最重要的复合数据类型之一，用于组合不同类型的字段

	// Address结构体：表示地址信息
	// 每个字段都有名称和类型，这是Go结构体的基本语法
	type Address struct {
		Street  string // 街道：字符串类型
		City    string // 城市：字符串类型
		Country string // 国家：字符串类型
		ZipCode string // 邮编：字符串类型
	}

	// Phone结构体：表示电话信息
	// Go中的字符串使用双引号包围，注释使用双斜杠
	type Phone struct {
		Type   string // 电话类型："mobile"手机，"home"家庭，"work"工作
		Number string // 电话号码
	}

	// User结构体：表示用户信息，展示了嵌套结构体的用法
	type User struct {
		ID       int                    // 用户ID：整数类型
		Name     string                 // 用户名：字符串类型
		Email    string                 // 邮箱：字符串类型
		IsActive bool                   // 是否活跃：布尔类型（true/false）
		Address  Address                // 嵌套结构体：Address类型，表示用户地址
		Phones   []Phone                // 嵌套切片：Phone类型的切片（数组），可以存储多个电话
		Tags     []string               // 字符串切片：存储多个标签
		Metadata map[string]interface{} // 动态字段：map类型，键是string，值是interface{}（任意类型）
		// map[string]interface{}是Go中实现灵活数据结构的常用方式
	}

	// Department结构体：表示部门信息
	type Department struct {
		ID          int    // 部门ID：整数类型
		Name        string // 部门名称：字符串类型
		Description string // 部门描述：字符串类型
		Users       []User // 用户数组：User类型的切片，一个部门可以有多个用户
	}

	// Company结构体：表示公司信息，是最顶层的结构
	type Company struct {
		ID          int          // 公司ID：整数类型
		Name        string       // 公司名称：字符串类型
		Website     string       // 公司网站：字符串类型
		Departments []Department // 部门数组：Department类型的切片，一个公司可以有多个部门
	}

	// 2. 声明一个复杂的多层级数据结构
	// var关键字用于声明变量，[]Company表示Company类型的切片
	// 这里声明了一个空的切片（nil切片），后续可以通过append函数添加元素
	var companies []Company
	// 语法说明：var 变量名 类型 = 初始值（这里省略了初始值，默认为零值）

	// 3. 创建并添加第一个公司
	// := 是Go的短变量声明语法，自动推断类型并赋值
	// Company{...} 是结构体的字面量语法，用于创建结构体实例
	company1 := Company{
		ID:      1,           // 字段名: 值，这是结构体初始化的语法
		Name:    "Tech Corp", // 字符串值使用双引号
		Website: "https://techcorp.com",
		// Departments字段是一个切片，使用[]Department{...}语法初始化
		Departments: []Department{
			// 第一个部门：Engineering部门
			{
				ID:          1,             // 部门ID
				Name:        "Engineering", // 部门名称
				Description: "软件开发部门",      // 部门描述
				// Users字段是一个User切片，包含多个用户
				Users: []User{
					// 第一个用户：张三
					{
						ID:       1,                       // 用户ID
						Name:     "张三",                    // 用户名
						Email:    "zhangsan@techcorp.com", // 邮箱
						IsActive: true,                    // 是否活跃
						// Address字段是一个嵌套的结构体
						Address: Address{
							Street:  "中关村大街1号", // 街道地址
							City:    "北京",      // 城市
							Country: "中国",      // 国家
							ZipCode: "100080",  // 邮编
						},
						// Phones字段是一个Phone切片
						Phones: []Phone{
							{Type: "mobile", Number: "13800138000"}, // 手机号
							{Type: "work", Number: "010-88888888"},  // 工作电话
						},
						// Tags字段是一个字符串切片
						Tags: []string{"高级工程师", "Go专家", "团队负责人"},
						// Metadata字段是一个map，键值对集合
						Metadata: map[string]interface{}{
							"joinDate":    "2020-01-15",                       // 入职日期：字符串
							"salary":      50000,                              // 薪资：整数
							"skills":      []string{"Go", "Python", "Docker"}, // 技能：字符串切片
							"isRemote":    true,                               // 是否远程工作：布尔值
							"performance": 9.5,                                // 绩效评分：浮点数
						},
					},
					// 第二个用户：李四
					{
						ID:       2,
						Name:     "李四",
						Email:    "lisi@techcorp.com",
						IsActive: true,
						Address: Address{
							Street:  "陆家嘴环路1000号",
							City:    "上海",
							Country: "中国",
							ZipCode: "200120",
						},
						Phones: []Phone{
							{Type: "mobile", Number: "13900139000"},
						},
						Tags: []string{"前端工程师", "React专家"},
						Metadata: map[string]interface{}{
							"joinDate":    "2021-03-20",
							"salary":      40000,
							"skills":      []string{"JavaScript", "React", "Vue"},
							"isRemote":    false,
							"performance": 8.8,
						},
					},
				},
			},
			// 第二个部门：Marketing部门
			{
				ID:          2,
				Name:        "Marketing",
				Description: "市场营销部门",
				Users: []User{
					// 第三个用户：王五
					{
						ID:       3,
						Name:     "王五",
						Email:    "wangwu@techcorp.com",
						IsActive: true,
						Address: Address{
							Street:  "天府大道北段1700号",
							City:    "成都",
							Country: "中国",
							ZipCode: "610041",
						},
						Phones: []Phone{
							{Type: "mobile", Number: "13700137000"},
							{Type: "home", Number: "028-88888888"},
						},
						Tags: []string{"市场经理", "品牌专家"},
						Metadata: map[string]interface{}{
							"joinDate":    "2019-07-10",
							"salary":      45000,
							"skills":      []string{"市场营销", "品牌管理", "数据分析"},
							"isRemote":    true,
							"performance": 9.2,
						},
					},
				},
			},
		},
	}

	// 4. 添加到公司数组中
	// append是Go的内建函数，用于向切片添加元素
	// 语法：append(切片, 元素1, 元素2, ...)
	// append会返回一个新的切片，所以需要重新赋值
	companies = append(companies, company1)

	// 5. 创建并添加第二个公司（使用不同的初始化方式）
	// 这次我们分步骤来构建复杂的数据结构，展示另一种初始化方式
	company2 := Company{
		ID:      2,
		Name:    "DataSoft",
		Website: "https://datasoft.io",
		// 注意：这里没有初始化Departments字段，它将是零值（nil切片）
	}

	// 先添加空部门，然后逐步填充
	// 这种方式适合动态构建复杂数据结构
	company2.Departments = append(company2.Departments, Department{
		ID:          1,
		Name:        "Data Science",
		Description: "数据科学部门",
		// 同样，Users字段暂时为零值
	})

	// 向第一个部门添加用户
	// 使用索引访问切片元素：切片[索引]
	company2.Departments[0].Users = append(company2.Departments[0].Users, User{
		ID:       4,
		Name:     "赵六",
		Email:    "zhaoliu@datasoft.io",
		IsActive: true,
		Address: Address{
			Street:  "深南大道10000号",
			City:    "深圳",
			Country: "中国",
			ZipCode: "518000",
		},
		Phones: []Phone{
			{Type: "mobile", Number: "13600136000"},
		},
		Tags: []string{"数据科学家", "机器学习专家"},
		Metadata: map[string]interface{}{
			"joinDate":    "2022-01-05",
			"salary":      60000,
			"skills":      []string{"Python", "TensorFlow", "PyTorch", "机器学习"},
			"isRemote":    false,
			"performance": 9.8,
		},
	})

	// 将第二个公司添加到companies切片中
	companies = append(companies, company2)

	// 6. 打印完整的多层级数据结构
	// fmt.Println是Go的标准输出函数，用于打印一行文本
	// fmt.Printf是格式化输出函数，可以按照指定格式输出变量值
	fmt.Println("=== 复杂多层级数组数据结构 ===")
	// %d是格式化占位符，表示整数；\n是换行符
	fmt.Printf("公司总数: %d\n\n", len(companies))
	// len()是Go的内建函数，用于获取切片、数组、字符串等的长度

	// 7. 遍历并展示多层级数据
	// for循环是Go中唯一的循环结构，有多种形式
	// 这里使用的是"for range"循环，用于遍历切片、数组、map等
	//
	// ====== range 关键字详解 ======
	// range 是Go语言中用于迭代（遍历）的关键字，可以遍历：
	// - 切片(slice)、数组(array)：返回索引和值
	// - 字符串(string)：返回字节位置和rune字符
	// - 映射(map)：返回键和值
	// - 通道(channel)：返回通道中的值
	//
	// 语法格式：
	// 1. for index, value := range collection { ... }  // 获取索引和值
	// 2. for index := range collection { ... }         // 只获取索引
	// 3. for _, value := range collection { ... }      // 只获取值，忽略索引（用下划线_）
	//
	// 在我们的代码中，companies是一个Company类型的切片，range会返回每个元素的索引和值

	for compIdx, company := range companies {
		// compIdx是索引（从0开始），company是当前遍历到的Company结构体值
		// 例如：第一次循环时，compIdx=0，company是第一个Company
		//       第二次循环时，compIdx=1，company是第二个Company
		// %d表示整数，%s表示字符串，compIdx+1使编号从1开始而不是0
		fmt.Printf("公司 %d: %s (%s)\n", compIdx+1, company.Name, company.Website)
		// 访问结构体字段使用点号：结构体变量.字段名
		fmt.Printf("  部门数量: %d\n", len(company.Departments))
		// 注意：缩进使用空格，不是Tab（Go代码风格要求）

		// 嵌套循环：遍历每个公司的部门
		// company.Departments是一个Department类型的切片
		// range会遍历这个切片中的每一个部门
		for deptIdx, department := range company.Departments {
			// deptIdx是部门在部门切片中的索引（从0开始）
			// department是当前遍历到的Department结构体值
			fmt.Printf("  部门 %d: %s - %s\n", deptIdx+1, department.Name, department.Description)
			fmt.Printf("    员工数量: %d\n", len(department.Users))

			// 三层嵌套循环：遍历每个部门的用户
			// department.Users是一个User类型的切片
			// range会遍历这个切片中的每一个用户
			for userIdx, user := range department.Users {
				// userIdx是用户在用户切片中的索引（从0开始）
				// user是当前遍历到的User结构体值
				// %t是布尔值的格式化占位符，输出true或false
				fmt.Printf("    员工 %d: %s (%s) - 活跃: %t\n",
					userIdx+1, user.Name, user.Email, user.IsActive)
				// 访问嵌套结构体的字段：user.Address.Street
				fmt.Printf("      地址: %s, %s, %s %s\n",
					user.Address.Street, user.Address.City, user.Address.Country, user.Address.ZipCode)

				// 遍历用户的电话列表
				fmt.Printf("      电话: ")
				// user.Phones是一个Phone类型的切片
				// 如果只需要值而不需要索引，可以使用下划线_忽略索引
				for phoneIdx, phone := range user.Phones {
					// phoneIdx是电话在电话切片中的索引（从0开始）
					// phone是当前遍历到的Phone结构体值
					// 条件语句：if 条件 { ... }
					if phoneIdx > 0 {
						// 如果不是第一个电话（索引大于0），添加逗号分隔
						fmt.Printf(", ")
					}
					// 输出电话号码和类型
					fmt.Printf("%s (%s)", phone.Number, phone.Type)
				}
				// 输出换行符
				fmt.Println()

				// %v是万能格式化占位符，可以输出任意类型的默认格式
				fmt.Printf("      标签: %v\n", user.Tags)

				// 遍历map（键值对集合）
				fmt.Printf("      元数据: ")
				// user.Metadata是一个map[string]interface{}类型的映射
				// for range遍历map时，第一个参数是键，第二个参数是值
				// 注意：遍历map的顺序是不确定的，每次运行可能不同
				for key, value := range user.Metadata {
					// key是map中的键（string类型）
					// value是对应的值（interface{}类型，可以是任意类型）
					// 输出键值对，用等号连接
					fmt.Printf("%s=%v ", key, value)
				}
				// 输出两个换行符，增加空行
				fmt.Println("\n")
			}
		}
		// 每个公司信息输出后添加空行
		fmt.Println()
	}

	// 8. 动态添加新数据到现有结构
	fmt.Println("=== 动态添加新数据 ===")

	// 向第一个公司的第一个部门添加新用户
	// 创建一个新的User结构体实例
	newUser := User{
		ID:       5,
		Name:     "孙七",
		Email:    "sunqi@techcorp.com",
		IsActive: true,
		Address: Address{
			Street:  "西湖区文三路",
			City:    "杭州",
			Country: "中国",
			ZipCode: "310000",
		},
		Phones: []Phone{
			{Type: "mobile", Number: "13500135000"},
		},
		Tags: []string{"实习工程师", "Go学习者"},
		Metadata: map[string]interface{}{
			"joinDate":    "2023-06-01",
			"salary":      15000,
			"skills":      []string{"Go", "基础编程"},
			"isRemote":    true,
			"performance": 7.5,
		},
	}

	// 使用索引访问切片元素：切片[索引]
	// companies[0]访问第一个公司
	// .Departments[0]访问第一个部门
	// .Users访问该部门的用户切片
	// append向切片添加新元素
	companies[0].Departments[0].Users = append(companies[0].Departments[0].Users, newUser)

	// 输出添加结果
	fmt.Printf("向 %s 的 %s 部门添加了新员工: %s\n",
		companies[0].Name, companies[0].Departments[0].Name, newUser.Name)
	fmt.Printf("该部门现在有 %d 名员工\n\n", len(companies[0].Departments[0].Users))

	// 9. 查询和过滤多层级数据
	fmt.Println("=== 数据查询和过滤 ===")

	// 查找所有活跃用户
	fmt.Println("所有活跃用户:")
	// 使用下划线_忽略索引，因为我们只需要值
	for _, company := range companies {
		for _, department := range company.Departments {
			for _, user := range department.Users {
				// if语句用于条件判断
				if user.IsActive {
					// 只有当user.IsActive为true时才会执行这里的代码
					fmt.Printf("- %s (%s) - %s - %s\n",
						user.Name, user.Email, company.Name, department.Name)
				}
			}
		}
	}

	// 查找特定技能的用户
	fmt.Println("\n具有 'Go' 技能的用户:")
	for _, company := range companies {
		for _, department := range company.Departments {
			for _, user := range department.Users {
				// 类型断言：将interface{}类型转换为具体的类型
				// 语法：值, ok := 变量.(目标类型)
				// ok是布尔值，表示转换是否成功
				if skills, ok := user.Metadata["skills"].([]string); ok {
					// 只有当转换成功且skills是[]string类型时才会执行这里的代码
					for _, skill := range skills {
						if skill == "Go" {
							fmt.Printf("- %s (%s) - %s\n", user.Name, user.Email, company.Name)
							// break跳出当前循环，不再检查该用户的其他技能
							break
						}
					}
				}
			}
		}
	}

	// 统计每个公司的员工数量
	fmt.Println("\n公司员工统计:")
	for _, company := range companies {
		// 声明并初始化变量：totalUsers := 0
		// 等价于：var totalUsers int = 0
		totalUsers := 0
		for _, department := range company.Departments {
			// +=是加法赋值运算符，等价于：totalUsers = totalUsers + len(department.Users)
			totalUsers += len(department.Users)
		}
		fmt.Printf("- %s: %d 名员工\n", company.Name, totalUsers)
	}
}
