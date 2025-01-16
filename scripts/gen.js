const path = require('path')
const fs = require('fs-extra')
const args = process.argv.slice(2)
const inquirer = require('inquirer').default

const apiVersion = 'v1'

// node scripts/gen.js ctx user/info
// node scripts/gen.js ctx dataset/version/delete/:id

// 是否确认
function isConfirmed(answersText) {
    return new Promise((resolve, reject) => {
        inquirer.prompt([{ type: 'confirm', name: 'ok', message: answersText, },
        ]).then(answers => {
            resolve(answers.ok)
        });
    })
}


async function main() {
    // 生成控制器
    if (args[0] === 'ctx') {
        let methodName = (args[2] || 'GET').toLocaleUpperCase()
        await genController(args[1], methodName)
    }
}

main()


// 生成控制器
async function genController(ctxPath, methodName) {
    // 控制器文件夹
    let ctxDirs = ctxPath.split('/')
    // 将路径中的中划线转为驼峰
    ctxDirs = ctxDirs.map(dir => {
        return dir.split('-').map((part, index) => {
            if (index === 0) return part;
            return part.charAt(0).toUpperCase() + part.slice(1);
        }).join('');
    });
    // 控制器函数名称
    let ctxHandName = [...ctxDirs].map(e => e.charAt(0).toUpperCase() + e.slice(1)).join('')
    // 控制器名称
    let ctxName = ctxDirs.pop()
    // 控制器名称 驼峰转下划线
    let _ctxName = ctxName.replace(/([A-Z])/g, '_$1').toLowerCase()
    // 控制器文件路径
    let ctxFilePath = path.join(__dirname, '../', 'internal/controller', ...ctxDirs, `${_ctxName}.go`)
    // 控制器API路径
    let routerPath = path.join('/api', apiVersion, ...ctxDirs, _ctxName.replace(/_/g, '-'))
    // 包名
    let ctxPkgName = 'ctx' + [...ctxDirs].map(e => e.charAt(0).toUpperCase() + e.slice(1)).join('')


    // 模板文件
    await fs.ensureDir(path.dirname(ctxFilePath))
    let ctxTemplate = fs.readFileSync(path.join(__dirname, 'template/ctx.tmp'), 'utf-8')
    ctxTemplate = ctxTemplate.replace('{{name}}', ctxHandName)
        .replace('{{pakname}}', ctxPkgName)
        .replace('{{action}}', routerPath)
    if (fs.existsSync(ctxFilePath)) {
        let isok = await isConfirmed('模板文件已存在，是否确认覆盖')
        if (isok) {
            fs.writeFileSync(ctxFilePath, ctxTemplate)
        }
    } else {
        fs.writeFileSync(ctxFilePath, ctxTemplate)
    }

    // 路由模板 ctxDirs, `${_ctxName}.go`
    let routerFilePath = path.join(__dirname, '../', 'internal/router', ...ctxDirs, `${_ctxName}.go`)
    await fs.ensureDir(path.dirname(routerFilePath))

    let routerTemplate = fs.readFileSync(path.join(__dirname, 'template/router.tmp'), 'utf-8')
    let routerPkgName = [...ctxDirs].map((e, i) => {
        if (i === 0) return e.charAt(0).toLowerCase() + e.slice(1)
        return e.charAt(0).toUpperCase() + e.slice(1)
    }).join('')

    routerTemplate = routerTemplate.replace(/{{pak}}/g, ctxPkgName)
        .replace('{{package}}', routerPkgName)
        .replace('{{path}}', routerPath)
        .replace('{{method}}', methodName)
        .replace('{{name}}', ctxHandName)
        .replace('{{ctxpath}}', ctxDirs.join('/'))

    if (fs.existsSync(routerFilePath)) {
        let isok = await isConfirmed('路由文件已存在，是否确认覆盖')
        if (isok) {
            fs.writeFileSync(routerFilePath, routerTemplate)
        }
    } else {
        fs.writeFileSync(routerFilePath, routerTemplate)
    }

    // 路由包路径
    const routerLibPath = `_ "${path.join("wp_template_display/internal/router/", ...ctxDirs)}"`
    var routerSetupText = fs.readFileSync(path.join(__dirname, '../', 'internal/setup/router.go'), 'utf-8')
    if (!routerSetupText.includes(routerLibPath)) {
        routerSetupText = routerSetupText.replace('// 导出路由', `// 导出路由\n\t${routerLibPath}`)
        fs.writeFileSync(path.join(__dirname, '../', 'internal/setup/router.go'), routerSetupText)
    }
}


// isConfirmed('文件夹已存在是否确认覆盖，').then(isok => {
//     console.log(isok)
// })